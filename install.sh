#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="supashift"
REPO_NAME="supashift"
BIN_NAME="supashift"

VERSION=""
BIN_DIR="${HOME}/.local/bin"
USE_SYSTEM=0
SKIP_CHECKSUM=0

usage() {
  cat <<USAGE
supashift installer

Usage:
  ./install.sh [options]

Options:
  --version <vX.Y.Z>   Install a specific version (default: latest release)
  --bin-dir <dir>      Install directory (default: ~/.local/bin)
  --system             Install to /usr/local/bin (uses sudo if needed)
  --skip-checksum      Skip SHA256 verification (not recommended)
  -h, --help           Show help
USAGE
}

log() {
  printf "[install] %s\n" "$*"
}

err() {
  printf "[install][error] %s\n" "$*" >&2
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

fetch() {
  local url="$1"
  local out="$2"
  if has_cmd curl; then
    curl -fsSL "$url" -o "$out"
  elif has_cmd wget; then
    wget -qO "$out" "$url"
  else
    err "Necesitas curl o wget"
    exit 1
  fi
}

fetch_text() {
  local url="$1"
  if has_cmd curl; then
    curl -fsSL "$url"
  elif has_cmd wget; then
    wget -qO- "$url"
  else
    err "Necesitas curl o wget"
    exit 1
  fi
}

sha256_file() {
  local file="$1"
  if has_cmd sha256sum; then
    sha256sum "$file" | awk '{print $1}'
  elif has_cmd shasum; then
    shasum -a 256 "$file" | awk '{print $1}'
  elif has_cmd openssl; then
    openssl dgst -sha256 "$file" | awk '{print $NF}'
  else
    err "No hay herramienta para SHA256 (sha256sum/shasum/openssl)"
    exit 1
  fi
}

need_tools() {
  local missing=0
  for t in tar mktemp; do
    if ! has_cmd "$t"; then
      err "Falta comando requerido: $t"
      missing=1
    fi
  done
  if [[ "$missing" -ne 0 ]]; then
    exit 1
  fi
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --version)
        VERSION="${2:-}"
        shift 2
        ;;
      --bin-dir)
        BIN_DIR="${2:-}"
        shift 2
        ;;
      --system)
        USE_SYSTEM=1
        shift
        ;;
      --skip-checksum)
        SKIP_CHECKSUM=1
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        err "Opción desconocida: $1"
        usage
        exit 1
        ;;
    esac
  done
}

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *)
      err "OS no soportado: $os"
      exit 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
      err "Arquitectura no soportada: $arch"
      exit 1
      ;;
  esac
}

resolve_version() {
  if [[ -n "$VERSION" ]]; then
    [[ "$VERSION" == v* ]] || VERSION="v${VERSION}"
    return
  fi

  local api url
  api="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
  log "Resolviendo última versión desde GitHub..."
  if ! has_cmd awk; then
    err "awk es requerido para resolver versión"
    exit 1
  fi
  url="$(fetch_text "$api")"
  VERSION="$(printf '%s' "$url" | awk -F '"' '/"tag_name":/ {print $4; exit}')"
  if [[ -z "$VERSION" ]]; then
    err "No se pudo resolver la última versión"
    exit 1
  fi
}

build_urls() {
  local ver_no_v
  ver_no_v="${VERSION#v}"
  ARCHIVE="${BIN_NAME}_${ver_no_v}_${OS}_${ARCH}.tar.gz"
  BASE="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}"
  ARCHIVE_URL="${BASE}/${ARCHIVE}"
  CHECKSUMS_URL="${BASE}/checksums.txt"
}

verify_checksum() {
  local archive_file checksums_file expected got
  archive_file="$1"
  checksums_file="$2"

  expected="$(awk -v f "${ARCHIVE}" '$2==f {print $1}' "$checksums_file")"
  if [[ -z "$expected" ]]; then
    err "No se encontró checksum para ${ARCHIVE}"
    exit 1
  fi

  got="$(sha256_file "$archive_file")"
  if [[ "$got" != "$expected" ]]; then
    err "Checksum inválido para ${ARCHIVE}"
    err "Esperado: $expected"
    err "Obtenido: $got"
    exit 1
  fi
  log "Checksum verificado"
}

install_binary() {
  local src_bin="$1"
  local target_dir="$BIN_DIR"

  if [[ "$USE_SYSTEM" -eq 1 ]]; then
    target_dir="/usr/local/bin"
  fi

  if [[ "$USE_SYSTEM" -eq 1 ]] && [[ ! -w "$target_dir" ]]; then
    if ! has_cmd sudo; then
      err "Necesitas sudo para instalar en $target_dir"
      exit 1
    fi
    sudo mkdir -p "$target_dir"
    sudo install -m 0755 "$src_bin" "$target_dir/${BIN_NAME}"
  else
    mkdir -p "$target_dir"
    install -m 0755 "$src_bin" "$target_dir/${BIN_NAME}"
  fi

  log "Instalado en ${target_dir}/${BIN_NAME}"
  if [[ ":$PATH:" != *":${target_dir}:"* ]]; then
    log "Tip: agrega ${target_dir} a tu PATH"
  fi
}

main() {
  parse_args "$@"
  need_tools
  detect_platform
  resolve_version
  build_urls

  log "Instalando ${BIN_NAME} ${VERSION} para ${OS}/${ARCH}"

  local tmpdir archive_file checksums_file
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  archive_file="${tmpdir}/${ARCHIVE}"
  checksums_file="${tmpdir}/checksums.txt"

  log "Descargando ${ARCHIVE_URL}"
  fetch "$ARCHIVE_URL" "$archive_file"

  if [[ "$SKIP_CHECKSUM" -eq 0 ]]; then
    log "Descargando ${CHECKSUMS_URL}"
    fetch "$CHECKSUMS_URL" "$checksums_file"
    verify_checksum "$archive_file" "$checksums_file"
  else
    log "Saltando validación de checksum"
  fi

  tar -xzf "$archive_file" -C "$tmpdir"
  if [[ ! -f "${tmpdir}/${BIN_NAME}" ]]; then
    err "No se encontró ${BIN_NAME} en el archivo descargado"
    exit 1
  fi

  install_binary "${tmpdir}/${BIN_NAME}"
  log "Listo. Ejecuta: ${BIN_NAME} --help"
}

main "$@"
