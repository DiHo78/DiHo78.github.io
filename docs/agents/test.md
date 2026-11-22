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
| Go tools | `go test ./...` (`tools/sitemapcheck/`, `tools/nothirdparty/`; fixtures in temp folders) | Go |
| Third-party requests | `go run ./tools/nothirdparty public https://diho78.github.io` (after the build; reads every .html, .svg, .css and .js file and fails on a script, stylesheet, font, image or frame from another origin; a URL a script builds at runtime stays unseen; the origin is the `baseURL` in `hugo.yaml`) | Go |
| Install script | `sh tools/install_test.sh` (folder checks of `tools/install.sh`; no download) | sh |
| Secrets | `gitleaks git --redact --no-banner --ignore-gitleaks-allow --log-opts=--all .` | gitleaks; full history (`fetch-depth: 0` in CI) |
