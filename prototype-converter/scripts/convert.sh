#!/bin/sh
# PROTOTYPE, throwaway: regenerate site/variants/* from the WordPress export.
# usage: sh scripts/convert.sh <export.xml>   (run from prototype-converter/; needs docker, curl, python3)
# The export holds commenters' e-mail and IP addresses: it is read here and never copied into the repo.
set -eu
XML=$(realpath "$1")
HERE=$(pwd)
TMP=$(mktemp -d)
POSTS="sql2k8r2-bpa-again es-muss-nicht-immer-dropcreate sql-server-2016-und-die-smos mein-kleiner-sqlsaturday-230"
WP2HUGO=1.37.0
BLOG2MD=67509ab
UA="agentos-consultant"

python3 -I scripts/filter.py "$XML" "$TMP/sample.xml" $POSTS

docker build -q -t proto-hugo:0.169.0 . >/dev/null

# wp2hugo: pinned release, checked against the release's checksums.txt
( cd "$TMP" \
  && curl -fsSL -A "$UA" -O "https://github.com/ashishb/wp2hugo/releases/download/$WP2HUGO/wp2hugo_Linux_x86_64.tar.gz" \
                       -O "https://github.com/ashishb/wp2hugo/releases/download/$WP2HUGO/checksums.txt" \
  && grep ' wp2hugo_Linux_x86_64.tar.gz$' checksums.txt | sha256sum -c - \
  && tar -xzf wp2hugo_Linux_x86_64.tar.gz wp2hugo )
mkdir -p "$TMP/wp2hugo-out"
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -v "$TMP":/work -w /work proto-hugo:0.169.0 \
  ./wp2hugo --source sample.xml --download-media --continue-on-media-download-error \
            --output wp2hugo-out --media-cache-dir /work/cache --color-log-output=false 2>&1 \
  | grep -v '"level":"debug"' > "$TMP/wp2hugo.log" || true
GEN=$(ls -d "$TMP"/wp2hugo-out/generated-*)

# blog2md: last commit, unmaintained since 2022; comments merged into the post (m) are none, filter.py drops them
mkdir -p "$TMP/blog2md"
curl -fsSL -A "$UA" "https://codeload.github.com/palaniraja/blog2md/tar.gz/$BLOG2MD" | tar -xz -C "$TMP/blog2md" --strip-components=1
docker run --rm -u "$(id -u):$(id -g)" -e HOME=/tmp -e npm_config_cache=/tmp/npm -v "$TMP":/work -w /work/blog2md \
  node:22-bookworm-slim sh -c 'npm ci --ignore-scripts --no-audit --no-fund >/dev/null 2>&1 && node index.js w /work/sample.xml /work/blog2md-out m paragraph-fix >/dev/null'

V=site/variants
rm -rf "$V"
mkdir -p "$V/raw-wp2hugo/content/posts" "$V/raw-wp2hugo/static" "$V/raw-blog2md/content/posts" "$V/raw-blog2md/static" "$V/fixed/static"
cp "$GEN"/content/posts/*.md "$V/raw-wp2hugo/content/posts/"
# an old Live Spaces permalink carries an authkey: keep it out of the repo
sed -i '/^spaces_[0-9a-f]*_permalink:/d' "$V"/raw-wp2hugo/content/posts/*.md
cp -r "$GEN/static/wp-content" "$V/raw-wp2hugo/static/"
cp "$TMP"/blog2md-out/*.md "$V/raw-blog2md/content/posts/"
python3 -I scripts/postprocess.py "$GEN" "$TMP/sample.xml" "$V/fixed/content"
touch "$V/raw-wp2hugo/static/.keep" "$V/raw-blog2md/static/.keep" "$V/fixed/static/.keep"
# Stack's archives widget warns (and --panicOnWarning fails) without a page of layout "archives"
for v in raw-wp2hugo raw-blog2md fixed; do
  mkdir -p "$V/$v/content/page/archives"
  printf -- '---\ntitle: Archives\nlayout: archives\nslug: archives\nmenu:\n  main:\n    weight: -70\n    params: {icon: archives}\n---\n' \
    > "$V/$v/content/page/archives/index.md"
done

# nothing personal may leave: no e-mail address in the generated content
if grep -rIlE '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[a-z]{2,}' "$V"; then echo "e-mail address found, see above" >&2; exit 1; fi
rm -rf "$TMP"
echo "done: $HERE/$V"
