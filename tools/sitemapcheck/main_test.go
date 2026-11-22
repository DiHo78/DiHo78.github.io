package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// site builds a site folder with an index.html under each page path.
func site(t *testing.T, pages ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, p := range pages {
		folder := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "index.html"), []byte("<html></html>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// sitemap writes a urlset with the given locations and returns its path.
func sitemap(t *testing.T, locs ...string) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, l := range locs {
		b.WriteString("<url><loc>" + l + "</loc></url>")
	}
	b.WriteString("</urlset>\n")
	return writeFile(t, b.String())
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sitemap.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func check(sitemapPath, siteDir string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run([]string{sitemapPath, siteDir}, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestAllPresent(t *testing.T) {
	sm := sitemap(t,
		"https://dirkhondong.wordpress.com/2024/01/30/die-bilder/",
		"https://dirkhondong.wordpress.com/2010/05/05/other-post/",
	)
	code, out, errOut := check(sm, site(t, "die-bilder", "other-post"))
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("code %d, stdout %q, stderr %q; want 0 and no output", code, out, errOut)
	}
}

func TestOneMissing(t *testing.T) {
	sm := sitemap(t,
		"https://dirkhondong.wordpress.com/2024/01/30/die-bilder/",
		"https://dirkhondong.wordpress.com/2010/05/05/gone/",
	)
	code, out, _ := check(sm, site(t, "die-bilder"))
	want := "https://dirkhondong.wordpress.com/2010/05/05/gone/ -> /gone/\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestEncodedSlug(t *testing.T) {
	sm := sitemap(t,
		"https://dirkhondong.wordpress.com/2010/08/25/ok-i%c2%b4m-the-unsteady-blogger%e2%80%a6/",
		"https://dirkhondong.wordpress.com/2010/05/05/catchin%e2%80%99-up/",
	)
	code, out, _ := check(sm, site(t, "ok-i´m-the-unsteady-blogger…"))
	want := "https://dirkhondong.wordpress.com/2010/05/05/catchin%e2%80%99-up/ -> /catchin’-up/\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestOwnPages(t *testing.T) {
	sm := sitemap(t,
		"https://dirkhondong.wordpress.com/impressum/",
		"https://dirkhondong.wordpress.com/datenschutz/",
		"https://dirkhondong.wordpress.com/about-2/",
	)
	code, out, _ := check(sm, site(t, "impressum", "datenschutz"))
	want := "https://dirkhondong.wordpress.com/about-2/ -> /about-2/\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestHomeURL(t *testing.T) {
	sm := sitemap(t,
		"https://dirkhondong.wordpress.com",
		"https://dirkhondong.wordpress.com/",
	)
	if code, out, _ := check(sm, site(t, "")); code != 0 || out != "" {
		t.Fatalf("with home page: code %d, stdout %q; want 0 and no output", code, out)
	}
	code, out, _ := check(sm, site(t))
	want := "https://dirkhondong.wordpress.com -> /\nhttps://dirkhondong.wordpress.com/ -> /\n"
	if code != 1 || out != want {
		t.Fatalf("without home page: code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestOtherShape(t *testing.T) {
	locs := []string{
		"https://dirkhondong.wordpress.com/category/sql-server/",
		"https://dirkhondong.wordpress.com/2010/05/05/",
		"https://dirkhondong.wordpress.com/2010/05/05/post",
		"https://dirkhondong.wordpress.com/2010/05/05/a%2fb/",
		"https://dirkhondong.wordpress.com/2010/05/05/../",
		"https://dirkhondong.wordpress.com/impressum/?p=1",
		"/impressum/",
		"",
	}
	// Every page a wrong mapping could find exists.
	sm := sitemap(t, locs...)
	code, out, _ := check(sm, site(t, "", "category/sql-server", "2010/05/05", "post", "a/b", "impressum"))
	var want strings.Builder
	for _, l := range locs {
		want.WriteString(l + " -> no page for this URL shape\n")
	}
	if code != 1 || out != want.String() {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want.String())
	}
}

func TestUnparsableSitemap(t *testing.T) {
	cases := map[string]string{
		"truncated":     `<?xml version="1.0"?><urlset><url><loc>https://dirkhondong.wordpress.com/impressum/</loc>`,
		"sitemap index": `<?xml version="1.0"?><sitemapindex><sitemap><loc>https://dirkhondong.wordpress.com/sitemap-1.xml</loc></sitemap></sitemapindex>`,
		"empty file":    ``,
		"no URL":        `<?xml version="1.0"?><urlset></urlset>`,
	}
	paths := map[string]string{"missing file": filepath.Join(t.TempDir(), "nope.xml")}
	for name, content := range cases {
		paths[name] = writeFile(t, content)
	}
	for name, path := range paths {
		t.Run(name, func(t *testing.T) {
			code, out, errOut := check(path, site(t, "impressum"))
			if code != 2 || out != "" || errOut == "" {
				t.Fatalf("code %d, stdout %q, stderr %q; want 2, no stdout and an error", code, out, errOut)
			}
		})
	}
}

func TestUsage(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"only-one"}, &out, &errOut); code != 2 || !strings.HasPrefix(errOut.String(), "usage:") {
		t.Fatalf("code %d, stderr %q; want 2 and a usage line", code, errOut.String())
	}
}

func TestSnapshotMapsEveryURL(t *testing.T) {
	code, out, errOut := check("wordpress-sitemap.xml", site(t))
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if code != 1 || errOut != "" || len(lines) != 56 {
		t.Fatalf("code %d, %d lines, stderr %q; want 1 and 56 lines", code, len(lines), errOut)
	}
	for _, l := range lines {
		if strings.HasSuffix(l, "no page for this URL shape") {
			t.Errorf("no mapping: %s", l)
		}
	}
	for _, want := range []string{
		" -> /",
		" -> /about-2/",
		" -> /ok-i´m-the-unsteady-blogger…/",
		" -> /catchin’-up-yesterday’s-issue…/",
	} {
		if !strings.Contains(out, want+"\n") {
			t.Errorf("no line ends with %q", want)
		}
	}
}
