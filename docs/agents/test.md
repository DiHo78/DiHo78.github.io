# Test suites

Install the tools once with `sh tools/install.sh <dir>`, into a new or empty
folder or an earlier install, then source `<dir>/env.sh`
in the same shell command. Run every suite from the repo root; CI runs them all.
None needs Docker.

| Suite | Command | Needs |
|---|---|---|
| Build | `hugo --panicOnWarning` | Hugo extended, Go, Dart Sass; network for the Stack module on the first build |
| Links | `htmltest` (after the build; internal links, images, anchors; `.htmltest.yml`) | htmltest |
| Lint | `markdownlint-cli2` (content; `.markdownlint-cli2.yaml`) | Node, markdownlint-cli2 |
| Go tools | `go test ./... && go -C tools/nothirdparty test ./...` (`tools/sitemapcheck/`; `tools/nothirdparty/` is its own module, so the root `go test ./...` skips it; fixtures in temp folders) | Go; network for `golang.org/x/net` on the first run |
| Third-party requests | `go -C tools/nothirdparty run . ../../public https://diho78.github.io` (after the build; reads .html and .svg files with an HTML parser and fails on a script, stylesheet, font, image or frame from another origin; in .css, .js and .mjs files, inline scripts, event handler attributes (`on*`) and styles, fails on every absolute or protocol-relative URL of another origin, comments included, unless `tools/nothirdparty/allowlist.txt` names it; plain links pass; the origin is the `baseURL` in `hugo.yaml`. It guards against accidental loads and does not see HTML or scripts written to fool it, URLs a script builds at runtime, `srcdoc` or `<meta http-equiv="Link">`) | Go; network for `golang.org/x/net` on the first run |
| Install script | `sh tools/install_test.sh` (folder checks of `tools/install.sh`; no download) | sh |
| Secrets | `gitleaks git --redact --no-banner --ignore-gitleaks-allow --log-opts=--all .` | gitleaks; full history (`fetch-depth: 0` in CI) |
