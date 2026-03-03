#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Uso: $0 <version-sin-v> <github-owner> <github-repo> [maintainer-name] [maintainer-email]" >&2
  exit 1
fi

VERSION="$1"
OWNER="$2"
REPO="$3"
MAINTAINER_NAME="${4:-$USER}"
MAINTAINER_EMAIL="${5:-$USER@localhost}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
AUR_DIR="$ROOT/packaging/aur/supashift-bin"
TMPL="$AUR_DIR/PKGBUILD.template"
OUT="$AUR_DIR/PKGBUILD"
SRCINFO="$AUR_DIR/.SRCINFO"

CHECKSUMS_URL="https://github.com/${OWNER}/${REPO}/releases/download/v${VERSION}/checksums.txt"
CHECKSUMS="$(curl -fsSL "$CHECKSUMS_URL")"

X86_FILE="supashift_v${VERSION}_linux_x86_64.tar.gz"
ARM_FILE="supashift_v${VERSION}_linux_arm64.tar.gz"

SHA_X86="$(awk -v f="$X86_FILE" '$2==f{print $1}' <<<"$CHECKSUMS")"
SHA_ARM="$(awk -v f="$ARM_FILE" '$2==f{print $1}' <<<"$CHECKSUMS")"

if [[ -z "$SHA_X86" || -z "$SHA_ARM" ]]; then
  echo "No se encontraron checksums para ${X86_FILE} o ${ARM_FILE} en $CHECKSUMS_URL" >&2
  exit 1
fi

sed \
  -e "s|{{PKGVER}}|$VERSION|g" \
  -e "s|{{GITHUB_OWNER}}|$OWNER|g" \
  -e "s|{{GITHUB_REPO}}|$REPO|g" \
  -e "s|{{SHA_X86_64}}|$SHA_X86|g" \
  -e "s|{{SHA_AARCH64}}|$SHA_ARM|g" \
  -e "s|{{MAINTAINER_NAME}}|$MAINTAINER_NAME|g" \
  -e "s|{{MAINTAINER_EMAIL}}|$MAINTAINER_EMAIL|g" \
  "$TMPL" > "$OUT"

(
  cd "$AUR_DIR"
  makepkg --printsrcinfo > "$SRCINFO"
)

echo "Generados:"
echo "- $OUT"
echo "- $SRCINFO"
