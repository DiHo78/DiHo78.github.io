# Converter comparison: result (prototype, 2026-11-22)

**Recommendation: wp2hugo 1.37.0 plus a post-processing script.** wp2hugo gets the slugs, dates, taxonomy
slugs, tweets and uploaded images right. What it misses, a script of 135 lines fixes (`scripts/postprocess.py`;
see it on `localhost:1315`). blog2md loses the slugs and the images, breaks the code blocks, and writes
commenters' e-mail addresses into its output; it is unmaintained since 2022. Do not use it.

Tested on four real posts (README, "Look at it"), with Hugo 0.169.0 and Stack v4.0.3.

## Converters, raw output

| Case | wp2hugo 1.37.0 (`:1313`) | blog2md 67509ab (`:1314`) |
|---|---|---|
| Slug, URL | WordPress slug kept, `…` too; `url:` keeps the old `/YYYY/MM/DD/<slug>/` | no slug: Hugo builds it from the title, entities included (`drophellipcreate`, `ruumlckblick`, `#8230.`) |
| Title | HTML entities stay (`DROP&hellip;CREATE` shows literally); 20 of 55 titles in the export have them | same |
| Date | ISO 8601 | RFC 1123 string, Hugo parses it |
| Categories, tags | separate, but as slugs (`sql-server`), not names | merged into `tags`; no categories at all |
| Language | not set | not set |
| Front matter | 5 to 17 WordPress keys per post (`_oembed_*`, `_publicize_*`, `guid`, a Live Spaces permalink with an `authkey`, removed from the committed copy) | minimal |
| Live Writer code (one `<pre>` per line) | one fence per line, line numbers kept | adjacent fences glued (```` `````` ````): invalid Markdown |
| Gist (bare URL on its own line) | left as plain text | left as plain text |
| Tweet | quote with link: as decision 13 wants | same |
| Uploaded images | downloaded to `static/wp-content/uploads/`, not page bundles; `?w=300` kept; full-size link targets not downloaded (broken link) | not downloaded, absolute `http://` hotlinks |
| Flickr images | hotlinked (`farmN.staticflickr.com`) | hotlinked |
| Internal links | relative, but old `/YYYY/MM/DD/<slug>/` | absolute to the old blog |
| Comments | into `data/comments.yaml` | into the post, **with each commenter's e-mail** in the link title; the approved check is broken (`"0"` counts as approved) |
| Needs | `hugo` on `PATH`; writes a PaperMod site around the content | Node, `paragraph-fix` flag, else paragraphs collapse |
| Licence, upkeep | Apache-2.0; README: free for personal use, commercial use needs a licence; releases monthly | no LICENSE file (`package.json`: ISC); last commit 2022-04 |

## Post-processing on top of wp2hugo (`:1315`)

`scripts/postprocess.py` does, per post:

- Front matter reduced to `title` (entities decoded), `date`, `url: /<slug>/`, `lang`, `categories`, `tags` with
  their display names from the export; "Uncategorized" dropped.
- Live Writer fences merged, line numbers removed, language guessed (`sql`, `powershell`, `text`).
- Gist URL → one fenced block per Gist file (GitHub API, language from the Gist) plus a link to the Gist.
- Page bundle: every image copied next to `index.md`, `?w=` removed, missing full-size images fetched; own Flickr
  photos fetched from `live.staticflickr.com` as a local preview, the link to the Flickr page kept.
- Internal links `/YYYY/MM/DD/<slug>/` → `/<slug>/`.
- `lang` guessed from German and English function words: right for all four posts.

Comments never reach either converter: `scripts/filter.py` drops them from the export first. The real conversion
should do the same; that is simpler than cleaning `data/comments.yaml`.

## Stack and Hugo: lessons for part 1

- **Stack is v4 now**: module `github.com/CaiJimmy/hugo-theme-stack/v4`, v4.0.3 (2026-05-25), needs Hugo
  extended ≥ 0.157. ADR 0001, #1 and #2 still say `/v3`.
- **`hugo --panicOnWarning` fails with Stack v4.0.3** on Hugo ≥ 0.153: Stack's `head/style.html` calls `toCSS`
  (libsass), which Hugo flags as deprecated. Fix: override `layouts/_partials/head/style.html` with
  `css.Sass (dict "transpiler" "dartsass" "silenceDeprecations" (slice "import" "global-builtin"))`, and put
  Dart Sass into the image and into CI. Those two deprecations are the minimum; `mixed-decls` must not be listed
  (Dart Sass warns that it is obsolete).
- **Archives widget**: without a page of `layout: archives`, Stack warns, and `--panicOnWarning` fails. Add the
  page (`content/page/archives/index.md`) or drop the widget.
- **Google Fonts**: Stack loads Lato from `fonts.googleapis.com` in `layouts/_partials/head/custom-font.html`.
  An empty override removes it; the CSS falls back to system fonts.
- **Language per post**: Stack's `baseof.html` takes `<html lang>` from the site locale, and the front matter
  `lang` alone changes nothing. Override `baseof.html` (one change:
  `lang="{{ .Params.lang | default .Site.Language.Locale }}"`). GPL-3.0 applies to that copy.
- **Slugs with `…`, `´`, `’`**: `slug:` with the `/:slug/` permalink drops those characters
  (`sql2k8r2-bpa-again-because-…`). `url: "/<slug>/"` in the front matter keeps them; the built folder holds the
  literal UTF-8 characters (checked for `…`, `´` and `’`).
- **Remote images**: Stack's image render hook calls `resources.GetRemote` on every remote image at build time.
  A blocked or slow host gives a warning, so `--panicOnWarning` fails. All images local (decision 15) avoids it.
- **`hugo server` in Hugo 0.169 serves from disk**: it writes `public/` into the mounted repo. Several servers on
  one folder overwrite each other. Use `--renderToMemory`, and keep `public/` in `.gitignore`.
- **Compose**: inside the container `hugo server --bind 0.0.0.0`; the port mapping `127.0.0.1:1313:1313` keeps it
  on localhost. Hugo keeps the module cache in `/tmp/hugo_cache` (not `/go/pkg/mod`); a named volume there saves
  the download on each start.

## Facts from the full export, for part 2

- Gists: 7, all bare URLs on their own line, in 3 posts (not 14 as counted from the public pages).
- Flickr images in content: 5 (`farm4`, `farm6`, `farm8`, `farm9.staticflickr.com`); `live.staticflickr.com`
  serves the same paths.
- Titles with HTML entities: 20 of the 55 published posts and pages.
- 135 attachment URLs in the export point to `dirkhondong.files.wordpress.com`; the post content uses
  `dirkhondong.wordpress.com/wp-content/uploads/`.

## Limits

- Four posts only. Not tested: the post with 148 `<code>` tags (`alle-datenbankbackups-aus-einem-ordner-wiederherstellen`),
  Gutenberg blocks (`die-bilder-der-sqlkonferenz-2023-sind-endlich-da`), the three pages, `[caption]`.
- The Live Writer merge joins every pair of fences that only blank lines separate; two real neighbouring code
  blocks would merge too.
- The `lang` guess and the code language guess are word lists; the content QA has to check them.
- Not checked: how GitHub Pages and `htmltest` handle a folder named with `…`.
- `farm4/6/8/9.staticflickr.com` and `dirkhondong.files.wordpress.com` are blocked in the agent VM; the prototype
  did not need them.
