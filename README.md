# blog.hondong.eu

Blog über SQL Server, Performance-Tuning, Troubleshooting und PowerShell.

Ziel des Projekts: den bestehenden WordPress.com-Blog mit allen Artikeln nach Hugo
umziehen und auf GitHub Pages unter `https://blog.hondong.eu` hosten. Eine Weiterleitung
gibt es nicht: Der alte Blog bleibt öffentlich und verweist auf die neue Adresse.

Entscheidungen: [`docs/adr/0001-blog-migration.md`](docs/adr/0001-blog-migration.md).

Prüfungen: `hugo --panicOnWarning` baut die Seite, `htmltest` prüft die gebaute Seite, `markdownlint-cli2`
ist der Linter für die Inhalte. `go run ./tools/sitemapcheck <sitemap.xml> <Ordner>` meldet jede URL der alten
Sitemap ohne Seite im gebauten Ordner; der Snapshot liegt in `tools/sitemapcheck/wordpress-sitemap.xml`.
`go -C tools/nothirdparty run . ../../public https://diho78.github.io` meldet jedes Skript, Stylesheet, jede Schrift,
jedes Bild und jeden Frame, den die gebaute Seite von einem fremden Origin lädt. `go test ./...` testet die Werkzeuge,
`go -C tools/nothirdparty test ./...` die Prüfung auf fremde Origins, die ein eigenes Go-Modul ist.
Installieren mit `sh tools/install.sh <Ordner>` in einen neuen oder leeren Ordner; die Befehle stehen in
[`docs/agents/test.md`](docs/agents/test.md).
