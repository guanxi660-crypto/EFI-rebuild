ARCHS ?= amd64 arm64
export ARCHS

.PHONY: all singbox cloudflared nezha clean

all:
	./build.sh all

singbox:
	./build.sh singbox

cloudflared:
	./build.sh cloudflared

nezha:
	./build.sh nezha

clean:
	rm -rf dist .build