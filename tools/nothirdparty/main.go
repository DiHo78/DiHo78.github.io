// Command nothirdparty reports every resource a built site loads from an
// origin other than its own.
//
// Usage, from this folder: go run . <site folder> <origin>
// It reads every .html and .svg file with an HTML parser and checks the
// attributes that load something. In .css and .js files, inline scripts,
// event handlers and styles, every absolute or protocol-relative URL of
// another origin fails, comments included, unless allowlist.txt names it.
// It prints one line per external URL and exits 1 when there is one, 2 on
// an error. It guards
// against accidental loads, not against HTML written to fool it, and a URL
// that a script puts together at runtime stays unseen.
package main

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

//go:embed allowlist.txt
var allowlistFile string

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

var (
	// An absolute URL with two slashes, or a protocol-relative one: two
	// slashes after a character that cannot be part of a path, then a host.
	// It ends at a character that rarely stands in a URL, e.g. * of a closing
	// comment.
	textURL     = regexp.MustCompile(`(?i)(?:\b[a-z][a-z0-9+.-]*:|(?:^|[^a-z0-9_.~/:%-]))//[a-z0-9][^\s"'` + "`" + `()<>\[\]{}\\,;*]*`)
	jsSlash     = regexp.MustCompile(`(?i)\\(?:/|u002f|x2f)`)
	leadSlashes = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9+.-]*:)?/{2,}`)
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: nothirdparty <site folder> <origin>")
		return 2
	}
	allowed, err := parseAllowlist(allowlistFile)
	if err != nil {
		fmt.Fprintln(stderr, "allowlist.txt:", err)
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
		if ext != ".html" && ext != ".svg" && ext != ".css" && ext != ".js" && ext != ".mjs" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var loads []load
		if ext == ".html" || ext == ".svg" {
			if ext == ".html" {
				pages++
			}
			if loads, err = htmlLoads(string(data)); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		} else {
			loads = textLoads(string(data))
		}
		rel, _ := filepath.Rel(dir, path)
		for _, l := range loads {
			if external(l.url, own) && !(l.inText && allowed[l.url]) {
				fmt.Fprintf(stdout, "%s: %s\n", filepath.ToSlash(rel), l.url)
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

// parseAllowlist reads lines of an exact URL and the reason it passes; blank
// lines and lines that start with # are skipped.
func parseAllowlist(text string) (map[string]bool, error) {
	allowed := map[string]bool{}
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("line %d: %q has no reason", i+1, line)
		}
		allowed[fields[0]] = true
	}
	return allowed, nil
}

// A load is a URL the page loads. inText marks a URL found in script or
// style text, where the allowlist applies.
type load struct {
	url    string
	inText bool
}

// htmlLoads parses a page or an SVG file and returns its loads in document
// order.
func htmlLoads(page string) ([]load, error) {
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		return nil, err
	}
	var loads []load
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		loads = append(loads, attrLoads(n)...)
		switch n.Data {
		case "script":
			if !dataBlock(n) {
				loads = append(loads, textLoads(childText(n))...)
			}
		case "style":
			loads = append(loads, textLoads(childText(n))...)
		}
	}
	return loads, nil
}

// attrLoads returns the URLs that an element's attributes load.
func attrLoads(n *html.Node) []load {
	rel := attrValue(n, "rel")
	var loads []load
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "src", "poster", "data", "background":
			loads = append(loads, load{url: a.Val})
		case "srcset", "imagesrcset":
			for _, candidate := range strings.Split(a.Val, ",") {
				if fields := strings.Fields(candidate); len(fields) > 0 {
					loads = append(loads, load{url: fields[0]})
				}
			}
		case "href":
			if n.Data != "a" && n.Data != "area" && !(n.Data == "link" && onlyReference(rel)) {
				loads = append(loads, load{url: a.Val})
			}
		case "style":
			loads = append(loads, textLoads(a.Val)...)
		default:
			// An event handler such as onerror is an inline script.
			if strings.HasPrefix(strings.ToLower(a.Key), "on") {
				loads = append(loads, textLoads(a.Val)...)
			}
		}
	}
	return loads
}

func attrValue(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Namespace == "" && strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func childText(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

// dataBlock reports whether a script element holds JSON data, which a browser
// does not run.
func dataBlock(n *html.Node) bool {
	switch strings.ToLower(strings.TrimSpace(attrValue(n, "type"))) {
	case "application/ld+json", "application/json":
		return true
	}
	return false
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

// textLoads returns every absolute or protocol-relative URL in script or
// style text, comments included, with an escaped slash read as a slash.
func textLoads(text string) []load {
	var loads []load
	for _, m := range textURL.FindAllString(jsSlash.ReplaceAllString(text, "/"), -1) {
		if i := strings.Index(m, "//"); i > 0 && !strings.HasSuffix(m[:i], ":") {
			m = m[i:] // drop the character before a protocol-relative URL
		}
		loads = append(loads, load{url: strings.TrimRight(m, ".:"), inText: true})
	}
	return loads
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
