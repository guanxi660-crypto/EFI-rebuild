#!/usr/bin/env bash
# 构建 sing-box / cloudflared / nezha 的 c-shared 动态库（linux/amd64 与 linux/arm64）。
# 用法：
#   ./build.sh              # 构建全部
#   ./build.sh singbox      # 仅构建 sing-box
#   ./build.sh cloudflared  # 仅构建 cloudflared
#   ./build.sh nezha        # 仅构建 nezha
#   ARCHS=amd64 ./build.sh  # 仅构建 amd64
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$ROOT/dist"
SRC="$ROOT/.build"
ARCHS="${ARCHS:-amd64 arm64}"
SINGBOX_TAGS="with_quic,with_utls,with_wireguard,with_clash_api,badlinkname,tfogo_checklinkname0"

mkdir -p "$OUT" "$SRC"

cc_for() {
  case "$1" in
    amd64) echo "gcc" ;;
    arm64) echo "aarch64-linux-gnu-gcc" ;;
    *) echo "gcc" ;;
  esac
}

build_singbox() {
  local arch
  echo "==> sing-box: resolving dependencies"
  (
    cd "$ROOT/go-wrapper/singbox"
    go mod tidy
  )
  for arch in $ARCHS; do
    echo "==> sing-box ($arch)"
    (
      cd "$ROOT/go-wrapper/singbox"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -tags "$SINGBOX_TAGS" \
        -o "$OUT/singbox_${arch}.so" .
    )
  done
}

build_cloudflared() {
  local src="$SRC/cloudflared" arch
  if [ ! -d "$src" ]; then
    git clone --depth=1 https://github.com/cloudflare/cloudflared.git "$src"
  fi
  python3 "$ROOT/scripts/patch.py" cloudflared "$src"
  cp "$ROOT/go-wrapper/cloudflared/export.go" "$src/cmd/cloudflared/export.go"
  for arch in $ARCHS; do
    echo "==> cloudflared ($arch)"
    (
      cd "$src/cmd/cloudflared"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -o "$OUT/cloudflared_${arch}.so" .
    )
  done
}

build_nezha() {
  local src="$SRC/agent" arch
  if [ ! -d "$src" ]; then
    git clone --depth=1 https://github.com/nezhahq/agent.git "$src"
  fi
  python3 "$ROOT/scripts/patch.py" nezha "$src"
  cp "$ROOT/go-wrapper/nezha/export.go" "$src/cmd/agent/export.go"
  for arch in $ARCHS; do
    echo "==> nezha ($arch)"
    (
      cd "$src/cmd/agent"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -o "$OUT/nezha_${arch}.so" .
    )
  done
}

case "${1:-all}" in
  singbox) build_singbox ;;
  cloudflared) build_cloudflared ;;
  nezha) build_nezha ;;
  all) build_singbox; build_cloudflared; build_nezha ;;
  *) echo "usage: $0 [singbox|cloudflared|nezha|all]"; exit 1 ;;
esac

echo "==> artifacts:"
ls -lh "$OUT"