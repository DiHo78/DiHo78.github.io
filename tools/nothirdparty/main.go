// Command nothirdparty reports every resource a built site loads from an
// origin other than its own.
//
// Usage: go run ./tools/nothirdparty <site folder> <origin>
// It reads every .html and .css file, prints one line per external URL and
// exits 1 when there is one, 2 on an error.
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
	comment = regexp.MustCompile(`(?s)<!--.*?-->`)
	tag     = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9:-]*)((?:[^>"']|"[^"]*"|'[^']*')*)>`)
	attr    = regexp.MustCompile(`([^\s"'>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+)))?`)

	scriptElement = regexp.MustCompile(`(?is)<script\b((?:[^>"']|"[^"]*"|'[^']*')*)>(.*?)</script>`)
	// An absolute or protocol-relative URL in a JS string literal.
	scriptURL   = regexp.MustCompile("[\"'`]((?:[a-zA-Z][a-zA-Z0-9+.-]*:)?//[^\"'`\\s]*)")
	jsSlash     = regexp.MustCompile(`(?i)\\(?:/|u002f|x2f)`)
	leadSlashes = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*:)?/{2,}`)

	styleElement = regexp.MustCompile(`(?is)<style\b[^>]*>(.*?)</style>`)
	cssURL       = regexp.MustCompile(`(?i)url\(\s*(?:"([^"]*)"|'([^']*)'|([^)\s"']*))\s*\)`)
	cssImport    = regexp.MustCompile(`(?i)@import\s+(?:"([^"]*)"|'([^']*)')`)
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
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".css" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var loads []string
		if ext == ".html" {
			pages++
			loads = htmlLoads(string(data))
		} else {
			loads = cssLoads(string(data))
		}
		rel, _ := filepath.Rel(dir, path)
		for _, u := range loads {
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

// htmlLoads returns the URLs a page loads: from the attributes that load a
// resource and from its CSS.
func htmlLoads(page string) []string {
	page = comment.ReplaceAllString(page, "")
	var urls []string
	for _, m := range styleElement.FindAllStringSubmatch(page, -1) {
		urls = append(urls, cssLoads(m[1])...)
	}
	for _, m := range scriptElement.FindAllStringSubmatch(page, -1) {
		urls = append(urls, scriptLoads(m[1], m[2])...)
	}
	for _, t := range tag.FindAllStringSubmatch(page, -1) {
		name := strings.ToLower(t[1])
		matches := attr.FindAllStringSubmatch(t[2], -1)
		rel := ""
		for _, a := range matches {
			if strings.EqualFold(a[1], "rel") {
				rel = a[2] + a[3] + a[4]
			}
		}
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

// cssLoads returns the URLs of url() and @import in CSS.
func cssLoads(css string) []string {
	var urls []string
	for _, re := range []*regexp.Regexp{cssURL, cssImport} {
		for _, m := range re.FindAllStringSubmatch(css, -1) {
			urls = append(urls, strings.Join(m[1:], ""))
		}
	}
	return urls
}

// scriptLoads returns the absolute and protocol-relative URL literals of an
// inline script; a JSON data block runs nothing and is skipped.
func scriptLoads(attrs, script string) []string {
	for _, a := range attr.FindAllStringSubmatch(attrs, -1) {
		if strings.EqualFold(a[1], "type") {
			switch strings.ToLower(strings.TrimSpace(a[2] + a[3] + a[4])) {
			case "application/ld+json", "application/json":
				return nil
			}
		}
	}
	var urls []string
	for _, m := range scriptURL.FindAllStringSubmatch(jsSlash.ReplaceAllString(script, "/"), -1) {
		urls = append(urls, m[1])
	}
	return urls
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
