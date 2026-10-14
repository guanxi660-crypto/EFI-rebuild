ARCHS ?= amd64 arm64
export ARCHS

.PHONY: all bot tools status clean

all:
	./build.sh all

bot:
	./build.sh bot

tools:
	./build.sh tools

status:
	./build.sh status

clean:
	rm -rf dist .build