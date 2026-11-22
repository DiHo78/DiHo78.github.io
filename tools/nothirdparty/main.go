// Command nothirdparty reports every resource a built site loads from an
// origin other than its own.
//
// Usage: go run ./tools/nothirdparty <site folder> <origin>
// It reads every .html, .svg, .css and .js file, prints one line per external
// URL and exits 1 when there is one, 2 on an error. It reads the files as
// they are: a URL that a script puts together at runtime stays unseen.
package main

import (
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

var (
	tagAt = regexp.MustCompile(`^<([a-zA-Z][a-zA-Z0-9:-]*)((?:[^>"']|"[^"]*"|'[^']*')*)>`)
	attr  = regexp.MustCompile(`([^\s"'>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+)))?`)

	cssURL    = regexp.MustCompile(`(?i)url\(\s*(?:"([^"]*)"|'([^']*)'|([^)\s"']*))\s*\)`)
	cssString = regexp.MustCompile(`"([^"]*)"|'([^']*)'`)

	// A string that names a host: another scheme, or two slashes.
	urlLike     = regexp.MustCompile(`(?i)^\s*(?:(?:https?|wss?|ftp):|[/\\]{2})`)
	jsSlash     = regexp.MustCompile(`(?i)\\(?:/|u002f|x2f)`)
	leadSlashes = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*:)?/{2,}`)
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: nothirdparty <site folder> <origin>")
		return 2
	}
	dir := args[0]
	own, err := url.Parse(args[1])
	if err != nil || own.Scheme == "" || own.Host == "" || strings.Trim(own.Path, "/") != "" {
		fmt.Fprintf(stderr, "origin must be scheme://host, e.g. https://diho78.github.io: %q\n", args[1])
		return 2
	}
	code, pages := 0, 0
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		read, ok := readers[filepath.Ext(path)]
		if !ok {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".html" {
			pages++
		}
		rel, _ := filepath.Rel(dir, path)
		for _, u := range read(string(data)) {
			if external(u, own) {
				fmt.Fprintf(stdout, "%s: %s\n", filepath.ToSlash(rel), u)
				code = 1
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	// A wrong folder would pass without checking anything.
	if pages == 0 {
		fmt.Fprintf(stderr, "%s: no HTML file\n", dir)
		return 2
	}
	return code
}

// readers maps a file extension to the function that returns its loads.
var readers = map[string]func(string) []string{
	".html": htmlLoads,
	".svg":  htmlLoads,
	".css":  cssLoads,
	".js":   jsLoads,
	".mjs":  jsLoads,
}

// htmlLoads returns the URLs a page loads. It reads the page from left to
// right as a browser does: a comment only between tags, and the text of
// script, style, textarea and title as raw text up to its end tag.
func htmlLoads(page string) []string {
	var urls []string
	for i := 0; i < len(page); {
		j := strings.IndexByte(page[i:], '<')
		if j < 0 {
			break
		}
		i += j
		rest := page[i:]
		if strings.HasPrefix(rest, "<!--") {
			i += commentLength(rest)
			continue
		}
		m := tagAt.FindStringSubmatch(rest)
		if m == nil {
			i++
			continue
		}
		i += len(m[0])
		name := strings.ToLower(m[1])
		urls = append(urls, tagLoads(name, m[2])...)
		switch name {
		case "script", "style", "textarea", "title":
			end := indexFold(page[i:], "</"+name)
			if end < 0 {
				end = len(page) - i
			}
			text := page[i : i+end]
			i += end
			if name == "script" {
				urls = append(urls, scriptLoads(m[2], text)...)
			} else if name == "style" {
				urls = append(urls, cssLoads(text)...)
			}
		}
	}
	return urls
}

// commentLength returns the length of the comment at the start of s, up to
// and including the first "-->" or "--!>"; "<!-->" and "<!--->" are empty
// comments. An unclosed comment runs to the end.
func commentLength(s string) int {
	for _, empty := range []string{"<!-->", "<!--->"} {
		if strings.HasPrefix(s, empty) {
			return len(empty)
		}
	}
	end := len(s)
	for _, closer := range []string{"-->", "--!>"} {
		if k := strings.Index(s[4:], closer); k >= 0 && 4+k+len(closer) < end {
			end = 4 + k + len(closer)
		}
	}
	return end
}

// indexFold returns the index of the first ASCII case-insensitive match of
// sub in s, or -1.
func indexFold(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if strings.EqualFold(s[i:i+len(sub)], sub) {
			return i
		}
	}
	return -1
}

// tagLoads returns the URLs that a tag's attributes load.
func tagLoads(name, attrs string) []string {
	matches := attr.FindAllStringSubmatch(attrs, -1)
	rel := ""
	for _, a := range matches {
		if strings.EqualFold(a[1], "rel") {
			rel = a[2] + a[3] + a[4]
		}
	}
	var urls []string
	for _, a := range matches {
		value := html.UnescapeString(a[2] + a[3] + a[4])
		switch strings.ToLower(a[1]) {
		case "src", "poster", "data", "background":
			urls = append(urls, value)
		case "srcset", "imagesrcset":
			for _, candidate := range strings.Split(value, ",") {
				if fields := strings.Fields(candidate); len(fields) > 0 {
					urls = append(urls, fields[0])
				}
			}
		case "href", "xlink:href":
			if name != "a" && name != "area" && !(name == "link" && onlyReference(rel)) {
				urls = append(urls, value)
			}
		case "style":
			urls = append(urls, cssLoads(value)...)
		}
	}
	return urls
}

// onlyReference reports whether every link type in rel only names a page and
// loads nothing, e.g. canonical.
func onlyReference(rel string) bool {
	types := strings.Fields(strings.ToLower(rel))
	for _, t := range types {
		switch t {
		case "canonical", "alternate", "author", "license", "me", "next", "prev", "search", "help", "bookmark":
		default:
			return false
		}
	}
	return len(types) > 0
}

// cssLoads returns the URLs of url() in CSS, and every quoted string that
// names a host, which covers @import and image-set().
func cssLoads(css string) []string {
	var urls []string
	for _, m := range cssURL.FindAllStringSubmatch(css, -1) {
		urls = append(urls, strings.Join(m[1:], ""))
	}
	for _, m := range cssString.FindAllStringSubmatch(cssURL.ReplaceAllString(css, ""), -1) {
		if s := m[1] + m[2]; urlLike.MatchString(s) {
			urls = append(urls, s)
		}
	}
	return urls
}

// scriptLoads returns the loads of an inline script; a JSON data block runs
// nothing and is skipped.
func scriptLoads(attrs, script string) []string {
	for _, a := range attr.FindAllStringSubmatch(attrs, -1) {
		if strings.EqualFold(a[1], "type") {
			switch strings.ToLower(strings.TrimSpace(a[2] + a[3] + a[4])) {
			case "application/ld+json", "application/json":
				return nil
			}
		}
	}
	return jsLoads(script)
}

// jsLoads returns the string and template literals of a script that name a
// host, with an escaped slash read as a slash. Comments and regular
// expression literals are skipped.
func jsLoads(src string) []string {
	var urls []string
	for _, s := range jsStrings(src) {
		if s = jsSlash.ReplaceAllString(s, "/"); urlLike.MatchString(s) {
			urls = append(urls, s)
		}
	}
	return urls
}

// jsStrings returns the raw contents of the string and template literals of
// a script. A slash starts a regular expression where an operand is
// expected: at the start, after an operator or punctuator, or after a keyword
// such as return.
func jsStrings(src string) []string {
	var out []string
	var last byte // the last character outside comments and whitespace
	word := ""    // the identifier or number that ends at last
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				return out
			}
			i += 2 + end + 2
			continue
		case strings.HasPrefix(src[i:], "//"):
			end := strings.IndexByte(src[i:], '\n')
			if end < 0 {
				return out
			}
			i += end
			continue
		case c == '"' || c == '\'' || c == '`':
			j := i + 1
			for j < len(src) && src[j] != c {
				if src[j] == '\\' {
					j++
				}
				j++
			}
			out = append(out, src[i+1:min(j, len(src))])
			i = j + 1
			last, word = c, ""
			continue
		case c == '/' && regexAllowed(last, word):
			j, class := i+1, false
			for j < len(src) && src[j] != '\n' && (class || src[j] != '/') {
				switch src[j] {
				case '\\':
					j++
				case '[':
					class = true
				case ']':
					class = false
				}
				j++
			}
			i = j + 1
			last, word = '/', ""
			continue
		}
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
		case isIdent(c):
			if !isIdent(last) {
				word = ""
			}
			word += string(c)
			last = c
		default:
			last, word = c, ""
		}
		i++
	}
	return out
}

func isIdent(c byte) bool {
	return c == '_' || c == '$' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func regexAllowed(last byte, word string) bool {
	if last == 0 || strings.IndexByte("(,=:[!&|?{};+-*%<>~^", last) >= 0 {
		return true
	}
	switch word {
	case "return", "typeof", "case", "do", "else", "in", "of", "new", "delete", "void", "throw", "yield", "await", "instanceof":
		return true
	}
	return false
}

// external reports whether a browser on own loads u from another origin. As
// a browser does, it drops tabs and newlines, reads a backslash as a slash
// and any run of two or more slashes as the start of a host.
func external(u string, own *url.URL) bool {
	u = strings.NewReplacer("\t", "", "\n", "", "\r", "", `\`, "/").Replace(strings.TrimSpace(u))
	u = leadSlashes.ReplaceAllString(u, "$1//")
	parsed, err := url.Parse(u)
	if err != nil {
		return true
	}
	if parsed.Scheme != "" && parsed.Scheme != own.Scheme {
		// Another scheme: a host even without slashes, e.g. http:cdn.example.com.
		switch parsed.Scheme {
		case "data", "blob", "about":
			return false
		}
		return true
	}
	return parsed.Host != "" && !strings.EqualFold(parsed.Host, own.Host)
}
