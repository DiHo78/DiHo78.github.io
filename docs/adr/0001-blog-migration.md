# ADR 0001: Move the blog from WordPress.com to Hugo on GitHub Pages

Status: accepted, 2026-11-21 (grilling session). Part 1, everything without the WordPress XML export, is specced
in #1. Part 2, conversion and switch day, gets its own spec once the export is there.

## Context

The blog `dirkhondong.wordpress.com` moves to `https://blog.hondong.eu`. Inventory of the old blog, from the public
sitemap and the post pages:

- 52 posts (2010–2024), pages `/impressum/`, `/datenschutz/`, `/about-2/`.
- Language mixed: 2010–2012 mostly English (~25 posts), from 2013 German.
- Permalinks `/YYYY/MM/DD/<slug>/`; some slugs carry encoded characters (`…`, `´`, `’`). All 52 slugs are unique;
  none collides with `impressum`, `datenschutz`, `about`, `about-2`, `category`, `tag`.
- 101 images under `dirkhondong.wordpress.com/wp-content/uploads/`, plus 3 from Flickr, 6 via `i0.wp.com`.
- Code: mostly embedded Gists (14 embeds, all from GitHub user DiHo78), plus Live-Writer `<pre>` in 5 posts.
- Embeds: 2 tweets and Flickr photos, in about 6 posts.
- 58 comments in 18 posts.
- Categories and tags, e.g. SQL Server 48, SQL Community 26, PowerShell 14, Uncategorized 14.
- The old Impressum cites the TMG, replaced by the DDG (§ 5) since May 2024. The old Datenschutz covers WordPress
  services (Jetpack, Gravatar, Google, cookies).

## Decisions

1. Hugo on GitHub Pages, no paid service.
2. `dev` builds and checks (CI), `main` publishes; the user merges `dev` into `main`.
3. Authoring: Markdown locally, sometimes the GitHub web editor; no CMS.
4. Impressum and Datenschutz pages from launch.
5. No tracking, analytics, external fonts or cookie banner. Embedded content from the old posts is checked and
   turned into links.
6. Build in parallel under `https://diho78.github.io`, check, then one switch day: Pages custom domain, IONOS CNAME
   `blog` → `diho78.github.io`, Enforce HTTPS.
7. Permalinks: short form `/:slug/`, slug identical to the WordPress slug. Internal links between posts are
   rewritten from `/YYYY/MM/DD/<slug>/` to `/<slug>/`.
8. No WordPress.com Site Redirect purchase.
9. After the switch the old blog stays public, with a pinned "moved to blog.hondong.eu" post, a text widget with the
   same notice on every page, and "discourage search engines" on (to verify that the free plan has it).
10. Language: one content tree, no translations, no `/en/` tree. Each post keeps its language, set in front matter.
    The user writes in English from now on, so the UI is English. Impressum and Datenschutz stay German; footer
    labels "Impressum" and "Privacy".
11. Comments: no old comments migrated, no comment system (Stack, `comments.enabled = false`). Under each post a
    feedback note with contact links: Mastodon and LinkedIn at launch, as on the user's GitHub profile; a dedicated
    mail alias later.
12. Gists become fenced code blocks with a language (`sql`, `powershell`) plus a link to the Gist.
13. Tweets become a quote with a link; Flickr becomes a link, with a local preview image only for the user's own
    photos.
14. Categories and tags keep the WordPress paths `/category/<x>/` and `/tag/<x>/`; "Uncategorized" is dropped;
    posts may be re-sorted.
15. All images are downloaded into the repo as page bundles next to the post.
16. No search at launch (Stack's built-in search or Pagefind later).
17. Theme: Stack (`github.com/CaiJimmy/hugo-theme-stack/v4`, GPL-3.0), as used by jpomfret/jpomfret.github.io.
    Final only after the prototype shows it with 3–4 real posts.
18. Theme as a Hugo module (needs Go).
19. Checks: `hugo --panicOnWarning`; `htmltest` for internal links, images and anchors (external links not on every
    push); `markdownlint-cli2` as linter; a tool that checks every old sitemap URL maps to an existing `/<slug>/`.
20. Impressum and Datenschutz: taken over and rewritten to current law (DDG § 5, GDPR, GitHub Pages as host, no
    cookies, no tracking). The user reviews and owns the legal text.
21. About page: at `/about/`, with `/about-2/` kept as an alias; rewritten in English; the user approves the text.
22. No PR preview deploys; CI uploads the built site as an artifact.
23. Local preview: a `compose.yaml` runs `hugo server` on `localhost:1313` with Go, Hugo and the theme in the image;
    nothing installed on the Windows host besides Docker. Same Hugo version as CI.
24. CV one-pager `/cv/`: a separate follow-up project after go-live.

## Open

- Converter: prototype on branch `prototype/converter` from `dev`, wp2hugo vs blog2md on the real XML export.
  Compare code blocks, images, slugs with special characters, front matter and Gists; show Stack with 3–4 real posts.
- The raw XML export holds commenters' e-mail and IP addresses: it is never committed (`.gitignore`, gitleaks in CI).
- Mail alias for the feedback note, later.

## Facts checked

- Hugo removed its migrations page on 2026-10-14 (gohugoio/hugoDocs commit "content: Remove tools section", issue
  #1573). Its last version lists blog2md (XML of free wordpress.com sites, keeps approved comments) and wp2hugo (Go
  CLI, keeps original URLs, code highlights, downloads media, converts shortcodes, Gutenberg blocks, Gists).
  wordpress-to-hugo-exporter and wordhugopress need plugins or database access.
- WordPress.com Site Redirect (support page "Redirect a WordPress.com site address") needs matching permalinks and
  points to `http://` only; the page names no price.

## Consequences

- Old links keep working on the old blog, which points readers to the new one; there is no redirect.
- The new site loads nothing from third parties, so the Datenschutz stays short and no cookie banner is needed.
