#!/usr/bin/env bash
# 构建 bot / tools / status 的 c-shared 动态库（linux/amd64 与 linux/arm64）。
# 用法：
#   ./build.sh              # 构建全部
#   ./build.sh bot          # 仅构建 bot
#   ./build.sh tools        # 仅构建 tools
#   ./build.sh status       # 仅构建 status
#   ARCHS=amd64 ./build.sh  # 仅构建 amd64
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$ROOT/dist"
SRC="$ROOT/.build"
ARCHS="${ARCHS:-amd64 arm64}"
BOT_TAGS="with_gvisor,with_quic,with_utls,with_wireguard,with_clash_api,badlinkname,tfogo_checklinkname0"

mkdir -p "$OUT" "$SRC"

cc_for() {
  case "$1" in
    amd64) echo "gcc" ;;
    arm64) echo "aarch64-linux-gnu-gcc" ;;
    *) echo "gcc" ;;
  esac
}

build_bot() {
  local arch
  echo "==> bot: resolving dependencies"
  (
    cd "$ROOT/go-wrapper/bot"
    go mod tidy
  )
  for arch in $ARCHS; do
    echo "==> bot ($arch)"
    (
      cd "$ROOT/go-wrapper/bot"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -tags "$BOT_TAGS" \
        -o "$OUT/bot_${arch}.so" .
    )
  done
}

build_tools() {
  local src="$SRC/cloudflared" arch
  if [ ! -d "$src" ]; then
    git clone --depth=1 https://github.com/cloudflare/cloudflared.git "$src"
  fi
  python3 "$ROOT/scripts/patch.py" tools "$src"
  cp "$ROOT/go-wrapper/tools/export.go" "$src/cmd/cloudflared/export.go"
  for arch in $ARCHS; do
    echo "==> tools ($arch)"
    (
      cd "$src/cmd/cloudflared"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -o "$OUT/tools_${arch}.so" .
    )
  done
}

build_status() {
  local src="$SRC/agent" arch
  if [ ! -d "$src" ]; then
    git clone --depth=1 https://github.com/nezhahq/agent.git "$src"
  fi
  python3 "$ROOT/scripts/patch.py" status "$src"
  cp "$ROOT/go-wrapper/status/export.go" "$src/cmd/agent/export.go"
  for arch in $ARCHS; do
    echo "==> status ($arch)"
    (
      cd "$src/cmd/agent"
      GOOS=linux GOARCH="$arch" CGO_ENABLED=1 CC="$(cc_for "$arch")" \
        go build -buildmode=c-shared -trimpath -ldflags="-s -w" \
        -o "$OUT/status_${arch}.so" .
    )
  done
}

case "${1:-all}" in
  bot) build_bot ;;
  tools) build_tools ;;
  status) build_status ;;
  all) build_bot; build_tools; build_status ;;
  *) echo "usage: $0 [bot|tools|status|all]"; exit 1 ;;
esac

echo "==> artifacts:"
ls -lh "$OUT"