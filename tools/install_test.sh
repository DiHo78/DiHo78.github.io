#!/bin/sh
# Tests the folder checks of tools/install.sh; downloads nothing.
# Usage: sh tools/install_test.sh
set -u
install="$(cd "$(dirname "$0")" && pwd)/install.sh"
root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT
failed=0

check() {
  if [ "$1" = 0 ]; then echo "ok   $2"; else echo "FAIL $2"; failed=1; fi
}

# A folder the script did not create keeps its contents.
mkdir -p "$root/home/go"
touch "$root/home/go/keep"
sh "$install" "$root/home" go >/dev/null 2>&1
[ $? -ne 0 ] && [ -f "$root/home/go/keep" ] && [ ! -e "$root/home/bin" ]
check $? "refuses a non-empty folder it did not create"

# A folder name with shell syntax never reaches env.sh.
for name in 'a$(echo x)' 'a`echo x`' 'a"b' "a'b" 'a;b'; do
  sh "$install" "$root/$name" nosuchtool >/dev/null 2>&1
  [ $? -ne 0 ] && [ ! -e "$root/$name" ]
  check $? "refuses the folder name $name"
done

# An empty or new folder is accepted and marked, so a later run may replace it.
sh "$install" "$root/new" nosuchtool 2>"$root/err" >/dev/null
grep -q 'unknown tool' "$root/err" && [ -f "$root/new/.installed-by-tools-install-sh" ]
check $? "accepts and marks a new folder"
sh "$install" "$root/new" nosuchtool 2>"$root/err" >/dev/null
grep -q 'unknown tool' "$root/err"
check $? "accepts a folder it marked before"

exit $failed
