#!/usr/bin/env bash
# Install plx from the latest GitHub release, then run `plx setup`.
# Usage: curl -fsSL https://raw.githubusercontent.com/xldplx/plx/main/install.sh | bash
set -euo pipefail

REPO="${PLX_REPO:-xldplx/plx}"
BIN_DIR="${PLX_INSTALL_DIR:-${HOME}/.local/bin}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "plx install: required command not found: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd tar
need_cmd uname
need_cmd mktemp
need_cmd install

os="$(uname -s)"
arch="$(uname -m)"
case "$os" in
  Linux) goos=Linux ;;
  Darwin) goos=Darwin ;;
  *)
    echo "plx install: unsupported OS: $os (use install.ps1 on Windows)" >&2
    exit 1
    ;;
esac
case "$arch" in
  x86_64|amd64) goarch=x86_64 ;;
  arm64|aarch64) goarch=arm64 ;;
  *)
    echo "plx install: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

asset="plx_${goos}_${goarch}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "plx install: resolving latest release for ${REPO}…"
# Follow the releases/latest redirect to learn the tag without the API.
redirect_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")"
tag="${redirect_url##*/}"
if [ -z "$tag" ] || [ "$tag" = "latest" ]; then
  echo "plx install: could not determine latest release tag" >&2
  exit 1
fi

download_url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
echo "plx install: downloading ${asset} (${tag})…"
curl -fsSL -o "${tmp}/${asset}" "$download_url"
tar -xzf "${tmp}/${asset}" -C "$tmp"

if [ ! -f "${tmp}/plx" ]; then
  echo "plx install: archive did not contain a plx binary" >&2
  exit 1
fi

mkdir -p "$BIN_DIR"
install -m 0755 "${tmp}/plx" "${BIN_DIR}/plx"

case ":${PATH}:" in
  *:"${BIN_DIR}":*) ;;
  *)
    echo "plx install: note: add ${BIN_DIR} to your PATH if 'plx' is not found" >&2
    ;;
esac

echo "plx install: installed ${BIN_DIR}/plx (${tag})"
echo "plx install: running plx setup…"
"${BIN_DIR}/plx" setup
echo "plx install: done"
