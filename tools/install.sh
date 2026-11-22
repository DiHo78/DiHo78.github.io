#!/bin/sh
# Installs the pinned build and test tools into <dir> and writes <dir>/env.sh.
# Usage: sh tools/install.sh <dir> [tool ...]
# Tools: go hugo sass htmltest gitleaks node markdownlint (default: all).
# <dir> is new, empty or an earlier install; its path holds only A-Z a-z 0-9 . _ / -.
# Every SHA-256 is copied from the release's checksums file (Dart Sass: the
# asset digest GitHub shows; it publishes no checksums file).
# Needs curl, tar, sha256sum. markdownlint needs node and installs the
# versions in package-lock.json.
set -eu

GO_VERSION=1.27.4
GO_SHA256=7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840
HUGO_VERSION=0.170.1
HUGO_SHA256=841a5462853aefc9942e4de1e8a5d75584b6a5181941e6ad3f1d745190e03cd8
SASS_VERSION=1.108.2
SASS_SHA256=32d2eb5b83cbc180fce9657ff3c57401a9006c59eccfe0bfdde0aa755a66d42f
HTMLTEST_VERSION=0.17.0
HTMLTEST_SHA256=775c597ee74899d6002cd2d93076f897f4ba68686bceabe2e5d72e84c57bc0fb
GITLEAKS_VERSION=8.30.1
GITLEAKS_SHA256=551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb
NODE_VERSION=26.13.1
NODE_SHA256=518674de2b621728a5e4486f5fb3b7d2900bb3cc4aa97969edaa463db99f9778

[ $# -ge 1 ] || { echo "usage: sh tools/install.sh <dir> [tool ...]" >&2; exit 2; }
# <dir> goes into env.sh, which gets sourced: allow no shell syntax in it.
case $1 in
  *[!A-Za-z0-9._/-]*) echo "refused: <dir> may hold only A-Z a-z 0-9 . _ / -: $1" >&2; exit 2 ;;
esac
# The script replaces folders in <dir>: only in one it created itself.
marker=.installed-by-tools-install-sh
if [ -d "$1" ] && [ ! -f "$1/$marker" ] && [ -n "$(ls -A "$1")" ]; then
  echo "refused: $1 is not empty and was not created by tools/install.sh" >&2; exit 2
fi
mkdir -p "$1/bin"
touch "$1/$marker"
dir=$(cd "$1" && pwd)
case $dir in
  *[!A-Za-z0-9._/-]*) echo "refused: the path of <dir> may hold only A-Z a-z 0-9 . _ / -: $dir" >&2; exit 2 ;;
esac
shift
[ $# -gt 0 ] || set -- go hugo sass htmltest gitleaks node markdownlint
repo=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# fetch <url> <sha256>: downloads to $tmp/archive and checks the hash.
fetch() {
  curl -fsSL -A agentos-installer -o "$tmp/archive" "$1"
  echo "$2  $tmp/archive" | sha256sum -c -
}

# unpack <subdir>: extracts $tmp/archive into the new folder $tmp/<subdir>.
unpack() {
  mkdir "$tmp/$1"
  tar -xzf "$tmp/archive" -C "$tmp/$1"
}

for tool; do
  case $tool in
    go)
      fetch "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" "$GO_SHA256"
      unpack go
      rm -rf "$dir/go"
      cp -R "$tmp/go/go" "$dir/go" ;;
    hugo)
      fetch "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-amd64.tar.gz" "$HUGO_SHA256"
      unpack hugo
      cp "$tmp/hugo/hugo" "$dir/bin/hugo" ;;
    sass)
      fetch "https://github.com/sass/dart-sass/releases/download/${SASS_VERSION}/dart-sass-${SASS_VERSION}-linux-x64.tar.gz" "$SASS_SHA256"
      unpack sass
      rm -rf "$dir/dart-sass"
      cp -R "$tmp/sass/dart-sass" "$dir/dart-sass" ;;
    htmltest)
      fetch "https://github.com/wjdp/htmltest/releases/download/v${HTMLTEST_VERSION}/htmltest_${HTMLTEST_VERSION}_linux_amd64.tar.gz" "$HTMLTEST_SHA256"
      unpack htmltest
      cp "$tmp/htmltest/htmltest" "$dir/bin/htmltest" ;;
    gitleaks)
      fetch "https://github.com/gitleaks/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_linux_x64.tar.gz" "$GITLEAKS_SHA256"
      unpack gitleaks
      cp "$tmp/gitleaks/gitleaks" "$dir/bin/gitleaks" ;;
    node)
      fetch "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.gz" "$NODE_SHA256"
      unpack node
      n="$tmp/node/node-v${NODE_VERSION}-linux-x64"
      # npm and npx are symlinks; the agents' host mount holds no symlinks.
      rm "$n/bin/npm" "$n/bin/npx"
      printf '#!/bin/sh\nexec "$(dirname "$0")/node" "$(dirname "$0")/../lib/node_modules/npm/bin/%s-cli.js" "$@"\n' npm > "$n/bin/npm"
      printf '#!/bin/sh\nexec "$(dirname "$0")/node" "$(dirname "$0")/../lib/node_modules/npm/bin/%s-cli.js" "$@"\n' npx > "$n/bin/npx"
      chmod +x "$n/bin/npm" "$n/bin/npx"
      rm -rf "$dir/node"
      cp -R "$n" "$dir/node" ;;
    markdownlint)
      mkdir "$tmp/md"
      cp "$repo/package.json" "$repo/package-lock.json" "$tmp/md/"
      (cd "$tmp/md" && PATH="$dir/node/bin:$PATH" npm ci --no-bin-links --ignore-scripts --no-audit --no-fund)
      rm -rf "$dir/markdownlint"
      cp -R "$tmp/md" "$dir/markdownlint"
      printf '#!/bin/sh\nexec node "$(dirname "$0")/../markdownlint/node_modules/markdownlint-cli2/markdownlint-cli2-bin.mjs" "$@"\n' > "$dir/bin/markdownlint-cli2"
      chmod +x "$dir/bin/markdownlint-cli2" ;;
    *)
      echo "unknown tool: $tool" >&2; exit 2 ;;
  esac
done

cat > "$dir/env.sh" <<EOF
# Written by tools/install.sh. Source it: . $dir/env.sh
export PATH="$dir/bin:$dir/go/bin:$dir/dart-sass:$dir/node/bin:\$PATH"
export GOTOOLCHAIN=local
EOF
