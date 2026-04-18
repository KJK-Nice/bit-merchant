#!/usr/bin/env bash
# Generates PWA icon rasters from static/pwa/icon.svg using rsvg-convert (librsvg).
# Requires: rsvg-convert  (brew install librsvg)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR/.."
SRC="$ROOT/static/pwa/icon.svg"
OUT="$ROOT/static/pwa/icons"

if ! command -v rsvg-convert &>/dev/null; then
  echo "Error: rsvg-convert not found. Install with: brew install librsvg" >&2
  exit 1
fi

mkdir -p "$OUT"

# Standard icons (exact crop, no padding)
rsvg-convert -w 192  -h 192  "$SRC" -o "$OUT/icon-192.png"
rsvg-convert -w 512  -h 512  "$SRC" -o "$OUT/icon-512.png"
rsvg-convert -w 180  -h 180  "$SRC" -o "$OUT/apple-touch-180.png"
rsvg-convert -w 32   -h 32   "$SRC" -o "$OUT/favicon-32.png"
rsvg-convert -w 16   -h 16   "$SRC" -o "$OUT/favicon-16.png"

# Maskable icon: wrap the SVG in a padded canvas so the icon occupies 80%
# of the area (safe zone = 10% margin on each side).
MASKABLE_SVG=$(cat <<'EOF'
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
  <rect width="512" height="512" fill="#000000"/>
  <svg x="51.2" y="51.2" width="409.6" height="409.6" viewBox="0 0 512 512">
    <rect width="512" height="512" rx="64" ry="64" fill="#000000"/>
    <text x="50%" y="50%" font-family="Arial, sans-serif" font-size="256" fill="#ffffff" text-anchor="middle" dy=".3em" font-weight="bold">BM</text>
  </svg>
</svg>
EOF
)

echo "$MASKABLE_SVG" | rsvg-convert -w 512 -h 512 -o "$OUT/icon-512-maskable.png"

echo "PWA icons generated in $OUT:"
ls -lh "$OUT"
