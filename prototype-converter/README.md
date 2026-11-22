# PROTOTYPE: converter comparison (throwaway, never merged)

Question: which converter turns the WordPress export into the posts ADR 0001 asks for, wp2hugo or blog2md, and
how does Stack look with real posts? Answer: [REPORT.md](REPORT.md).

## Look at it

Needs Docker only. Clone the branch outside the agents' root folder, then:

```sh
cd prototype-converter
docker compose up
```

| URL | Content |
|---|---|
| <http://localhost:1313> | wp2hugo 1.37.0, raw output |
| <http://localhost:1314> | blog2md 67509ab, raw output (`m paragraph-fix`) |
| <http://localhost:1315> | wp2hugo + `scripts/postprocess.py`: the target shape |

The first start builds the image and fetches the Stack module (a minute or two). `Ctrl+C` stops it.
The raw variants hotlink images from WordPress and Flickr; they load only with internet access.

Four posts, chosen for the hard cases:

| Post | Case |
|---|---|
| `sql2k8r2-bpa-again…-because-i-was-a-little-bit-wrong` (en, 2010) | `…` in the slug, uploaded images with `?w=300` |
| `es-muss-nicht-immer-dropcreate-sein` (de, 2015) | Live Writer code: 23 `<pre>`, one per line, line numbers |
| `sql-server-2016-und-die-smos-wenn-version-130-nicht-so-recht-will` (de, 2018) | two Gists as bare URLs, images |
| `mein-kleiner-sqlsaturday-230-rckblick` (de, 2013) | tweet, own Flickr photos, internal links |

## Regenerate

`site/variants/` is committed so that `docker compose up` needs no export. To rebuild it from the export
(needs Docker, curl, python3; the export is read, never copied into the repo):

```sh
sh scripts/convert.sh <path to export.xml>
```

`scripts/filter.py` keeps only the four posts and their attachments and drops every comment and the author's
e-mail before either converter sees the file.

## Layout

- `Dockerfile`: Hugo extended 0.169.0, Go 1.25, Dart Sass 1.107.1, each checked against a SHA-256.
- `site/`: one Stack v4.0.3 site; `config/<variant>/hugo.yaml` picks the content folder per variant.
- `site/layouts/`: three Stack overrides the real site needs too (see REPORT.md, "Stack").
