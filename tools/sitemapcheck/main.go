// Command sitemapcheck reports every URL of the old WordPress sitemap that has
// no page in a built site folder.
//
// Usage: go run ./tools/sitemapcheck <sitemap.xml> <site folder>
// It prints one line per missing URL and exits 1 when one is missing, 2 when
// the sitemap cannot be read or parsed.
package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

var postPath = regexp.MustCompile(`^/[0-9]{4}/[0-9]{2}/[0-9]{2}/([^/]+)/$`)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: sitemapcheck <sitemap.xml> <site folder>")
		return 2
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	var set urlset
	if err := xml.Unmarshal(data, &set); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", args[0], err)
		return 2
	}
	// An empty urlset would pass without checking anything.
	if len(set.URLs) == 0 {
		fmt.Fprintf(stderr, "%s: no URL in the sitemap\n", args[0])
		return 2
	}
	code := 0
	for _, u := range set.URLs {
		page, ok := newPath(u.Loc)
		if !ok {
			fmt.Fprintf(stdout, "%s -> no page for this URL shape\n", u.Loc)
			code = 1
			continue
		}
		if _, err := os.Stat(filepath.Join(args[1], filepath.FromSlash(page), "index.html")); err != nil {
			fmt.Fprintf(stdout, "%s -> %s\n", u.Loc, page)
			code = 1
		}
	}
	return code
}

// newPath returns the path of the page on the new site for an old URL, and
// false for a URL of any other shape.
func newPath(loc string) (string, bool) {
	u, err := url.Parse(loc)
	if err != nil || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
		return "", false
	}
	switch u.Path {
	case "", "/":
		return "/", true
	case "/impressum/", "/datenschutz/", "/about-2/":
		return u.Path, true
	}
	m := postPath.FindStringSubmatch(u.Path)
	if m == nil || m[1] == "." || m[1] == ".." {
		return "", false
	}
	return "/" + m[1] + "/", true
}
