#!/usr/bin/env bash
# Puts together the course website for GitHub Pages, in the folder given (default _site):
#   site/                    the website itself
#   phase-*/day-*/README.md  the lessons, which site/day.html fetches as ../<day path>/README.md
#   index.html               sends the address with no path to site/
# That's what netlify.toml and vercel.json serve too, minus the rest of the repository.
# Every link is relative, so it works under https://<owner>.github.io/<repo>/. Try it locally:
#   tools/build-pages.sh /tmp/pages/100-Days-of-Go && python3 -m http.server -d /tmp/pages 8000
#   then open http://localhost:8000/100-Days-of-Go/
set -euo pipefail

cd "$(dirname "$0")/.."
out="${1:-_site}"

rm -rf "$out"
mkdir -p "$out"
cp -R site "$out/site"
count=0
for readme in phase-*/day-*/README.md; do
  mkdir -p "$out/$(dirname "$readme")"
  cp "$readme" "$out/$readme"
  count=$((count + 1))
done
cat > "$out/index.html" <<'HTML'
<!doctype html>
<meta charset="utf-8">
<title>100 Days of Go</title>
<meta http-equiv="refresh" content="0; url=site/">
<link rel="canonical" href="site/">
<p><a href="site/">100 Days of Go</a></p>
HTML

echo "Built $out: the website and $count lessons"
