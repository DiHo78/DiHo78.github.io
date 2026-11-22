package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const origin = "https://diho78.github.io"

// site builds a site folder from file paths and contents.
func site(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func check(dir string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run([]string{dir, origin}, &out, &errOut)
	return code, out.String(), errOut.String()
}

func page(head, body string) string {
	return "<!DOCTYPE html><html><head>" + head + "</head><body>" + body + "</body></html>"
}

func TestExternalScriptFails(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html":      page(`<script src="/js/main.js"></script>`, "<p>home</p>"),
		"post/index.html": page(`<script src="https://cdn.example.com/x.js"></script>`, ""),
	})
	code, out, _ := check(dir)
	want := "post/index.html: https://cdn.example.com/x.js\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestPageWithoutExternalLoadsPasses(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html": page(`<script src="/js/main.js"></script><link rel="stylesheet" href="/css/style.css">`, `<img src="img/a.png">`),
	})
	code, out, errOut := check(dir)
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("code %d, stdout %q, stderr %q; want 0 and no output", code, out, errOut)
	}
}

func TestExternalLoadsFail(t *testing.T) {
	cases := map[string]struct{ head, body, url string }{
		"stylesheet":            {`<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Lato">`, "", "https://fonts.googleapis.com/css2?family=Lato"},
		"preconnect":            {`<link rel=preconnect href=https://fonts.gstatic.com crossorigin>`, "", "https://fonts.gstatic.com"},
		"icon":                  {`<LINK REL='icon' HREF='https://cdn.example.com/favicon.ico'>`, "", "https://cdn.example.com/favicon.ico"},
		"image":                 {"", `<img alt="a > b" src="https://i0.wp.com/a.png">`, "https://i0.wp.com/a.png"},
		"srcset":                {"", `<img src="/a.png" srcset="/a-2x.png 2x, https://i0.wp.com/a-3x.png 3x">`, "https://i0.wp.com/a-3x.png"},
		"picture source":        {"", `<picture><source srcset="https://i0.wp.com/a.webp"><img src="/a.png"></picture>`, "https://i0.wp.com/a.webp"},
		"frame":                 {"", `<iframe src="https://www.youtube.com/embed/x"></iframe>`, "https://www.youtube.com/embed/x"},
		"video poster":          {"", `<video poster="https://cdn.example.com/p.jpg"></video>`, "https://cdn.example.com/p.jpg"},
		"object":                {"", `<object data="https://cdn.example.com/x.svg"></object>`, "https://cdn.example.com/x.svg"},
		"base":                  {`<base href="https://cdn.example.com/">`, "", "https://cdn.example.com/"},
		"svg use":               {"", `<svg><use xlink:href="https://cdn.example.com/s.svg#i"/></svg>`, "https://cdn.example.com/s.svg#i"},
		"entity in URL":         {"", `<img src="https:&#x2F;&#x2F;cdn.example.com/a.png">`, "https://cdn.example.com/a.png"},
		"protocol-relative":     {"", `<script src="//cdn.example.com/x.js"></script>`, "//cdn.example.com/x.js"},
		"style attribute":       {"", `<div style="background: url('https://cdn.example.com/bg.png')"></div>`, "https://cdn.example.com/bg.png"},
		"style element":         {`<style>@font-face { font-family: Lato; src: url(https://fonts.gstatic.com/lato.woff2) format("woff2"); }</style>`, "", "https://fonts.gstatic.com/lato.woff2"},
		"inline script import":  {"", `<script type="module">const m = await import("https:\/\/cdn.jsdelivr.net\/npm\/photoswipe@5.4.4\/dist\/photoswipe.esm.min.js");</script>`, "https://cdn.jsdelivr.net/npm/photoswipe@5.4.4/dist/photoswipe.esm.min.js"},
		"inline script string":  {"", `<script>link.href = '//cdn.example.com/x.css';</script>`, "//cdn.example.com/x.css"},
		"other scheme no slash": {"", `<script src="http:cdn.example.com/x.js"></script>`, "http:cdn.example.com/x.js"},
		"backslashes":           {"", `<script src="\\cdn.example.com/x.js"></script>`, `\\cdn.example.com/x.js`},
		"slash backslash":       {"", `<script src="/\cdn.example.com/x.js"></script>`, `/\cdn.example.com/x.js`},
		"three slashes":         {"", `<script src="///cdn.example.com/x.js"></script>`, "///cdn.example.com/x.js"},
		"background attribute":  {"", `<table background="https://cdn.example.com/b.png"></table>`, "https://cdn.example.com/b.png"},
		"style @import":         {`<style>@import "https://fonts.googleapis.com/css?family=Lato";</style>`, "", "https://fonts.googleapis.com/css?family=Lato"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			code, out, _ := check(site(t, map[string]string{"index.html": page(c.head, c.body)}))
			want := "index.html: " + c.url + "\n"
			if code != 1 || out != want {
				t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
			}
		})
	}
}

func TestLinksAndOwnOriginPass(t *testing.T) {
	cases := map[string]struct{ head, body string }{
		"plain links":                {"", `<a href="https://gohugo.io/" target="_blank" rel="noopener">Hugo</a> <a href="https://jimmycai.com">Jimmy</a> <a href="https://github.com/CaiJimmy/hugo-theme-stack">Stack</a>`},
		"canonical and alternate":    {`<link rel="canonical" href="https://diho78.github.io/"><link rel="alternate" type="application/rss+xml" href="https://diho78.github.io/index.xml"><link rel="me" href="https://mastodon.social/@x">`, ""},
		"own origin absolute":        {`<script src="https://diho78.github.io/js/main.js"></script>`, `<img src="https://diho78.github.io/a.png">`},
		"own origin protocol-rel":    {`<link rel="stylesheet" href="//diho78.github.io/style.css">`, ""},
		"data URL":                   {"", `<img src="data:image/png;base64,iVBORw0KGgo=">`},
		"commented out":              {`<!-- <script src="https://cdn.example.com/x.js"></script> -->`, ""},
		"text that looks like a URL": {"", `<p>src="https://cdn.example.com/x.js"</p>`},
		"about blank frame":          {"", `<iframe src="about:blank"></iframe>`},
		"JSON-LD":                    {`<script type="application/ld+json">{"@context": "https://schema.org"}</script>`, ""},
		"own scheme relative":        {"", `<img src="https:a.png"><img src="https:/a.png">`},
		"inline script own origin":   {"", `<script>fetch("https:\/\/diho78.github.io\/index.json"); const p = "/js/x.js";</script>`},
		"tab in own path":            {"", "<img src=\"/a\t.png\">"},
		"meta content":               {`<meta property="og:image" content="https://cdn.example.com/a.png">`, ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			code, out, errOut := check(site(t, map[string]string{"index.html": page(c.head, c.body)}))
			if code != 0 || out != "" || errOut != "" {
				t.Fatalf("code %d, stdout %q, stderr %q; want 0 and no output", code, out, errOut)
			}
		})
	}
}

func TestStylesheetFileFails(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html": page(`<link rel="stylesheet" href="/scss/style.css">`, ""),
		"scss/style.css": `@import url("https://fonts.googleapis.com/css2?family=Lato");` +
			`@font-face{font-family:Lato;src:url(/fonts/lato.woff2) format("woff2"),url('https://fonts.gstatic.com/lato.woff') format("woff")}` +
			`body{background:url(data:image/png;base64,iVBORw0KGgo=)}`,
	})
	code, out, _ := check(dir)
	want := "scss/style.css: https://fonts.googleapis.com/css2?family=Lato\nscss/style.css: https://fonts.gstatic.com/lato.woff\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestErrors(t *testing.T) {
	withPage := site(t, map[string]string{"index.html": page("", "")})
	cases := map[string][]string{
		"one argument":        {withPage},
		"missing folder":      {filepath.Join(t.TempDir(), "nope"), origin},
		"no HTML file":        {site(t, map[string]string{"style.css": "body{}"}), origin},
		"origin with no host": {withPage, "diho78.github.io"},
		"origin with a path":  {withPage, "https://diho78.github.io/blog/"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := run(args, &out, &errOut); code != 2 || out.Len() != 0 || errOut.Len() == 0 {
				t.Fatalf("code %d, stdout %q, stderr %q; want 2, no stdout and an error", code, out.String(), errOut.String())
			}
		})
	}
}

func TestScriptFileFails(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html": page(`<script type="module" src="/js/app.js"></script>`, ""),
		"js/app.js": `/*! Lib 1.0 - https://lib.example.com */
// see https://docs.example.com/x
const re = /["']/g, path = "a//b", half = 1 / 2;
import("https://cdn.example.com/x.js"); s.src = "https:\/\/cdn.example.com\/y.js";
const t = ` + "`//cdn.example.com/z.js`" + `; const own = "/js/other.js", own2 = "https://diho78.github.io/a.js";
const scheme = 'http:cdn.example.com/w.js'; // done
`,
	})
	code, out, _ := check(dir)
	want := "js/app.js: https://cdn.example.com/x.js\njs/app.js: https://cdn.example.com/y.js\njs/app.js: //cdn.example.com/z.js\njs/app.js: http:cdn.example.com/w.js\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestScriptCommentsPass(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html": page(`<script>/* https://cdn.example.com/a.js */ // https://cdn.example.com/b.js
var note = "Note: x";</script>`, ""),
		"js/lib.js": "/*!\n  * Lib 5.4.4 - https://lib.example.com\n  * (c) 2024\n  */\nfunction t(a){return a.replace(/\\/\\//g,\"/\")}export{t as default};\n",
	})
	code, out, errOut := check(dir)
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("code %d, stdout %q, stderr %q; want 0 and no output", code, out, errOut)
	}
}

func TestImageSetFails(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html":     page(`<style>body{background-image:image-set("https://cdn.example.com/a.png" 1x)}</style>`, `<div style="background-image: image-set('https://cdn.example.com/b.png' 2x)"></div>`),
		"scss/style.css": `.x{content:"Note: x";font-family:"Lato"}.y{background-image:image-set("//cdn.example.com/c.png" 1x, url(/d.png) 2x)}`,
	})
	code, out, _ := check(dir)
	want := "index.html: https://cdn.example.com/a.png\nindex.html: https://cdn.example.com/b.png\nscss/style.css: //cdn.example.com/c.png\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}

func TestCommentOpenerOutsideCommentFails(t *testing.T) {
	cases := map[string]string{
		"in an attribute": `<img alt="<!--" src="/a.png"><script src="https://cdn.example.com/x.js"></script><p title="-->"></p>`,
		"in script text":  `<script>var a="<!--";</script><script src="https://cdn.example.com/x.js"></script><script>var b="-->";</script>`,
		"in a title":      `<title><!--</title><script src="https://cdn.example.com/x.js"></script><p>--></p>`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			code, out, _ := check(site(t, map[string]string{"index.html": page("", body)}))
			want := "index.html: https://cdn.example.com/x.js\n"
			if code != 1 || out != want {
				t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
			}
		})
	}
}

func TestSVGFileFails(t *testing.T) {
	dir := site(t, map[string]string{
		"index.html":  page("", `<object data="/img/a.svg"></object>`),
		"img/a.svg":   `<svg xmlns="http://www.w3.org/2000/svg"><image href="https://cdn.example.com/a.png"/></svg>`,
		"img/own.svg": `<svg xmlns="http://www.w3.org/2000/svg"><use href="#i"/></svg>`,
	})
	code, out, _ := check(dir)
	want := "img/a.svg: https://cdn.example.com/a.png\n"
	if code != 1 || out != want {
		t.Fatalf("code %d, stdout %q; want 1 and %q", code, out, want)
	}
}
