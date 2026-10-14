#!/usr/bin/env python3
"""FFI 冒烟测试：用 ctypes 加载编译出的 .so 并校验导出的 C 接口。

用法：
    python3 tests/ffi_smoke.py dist/singbox_amd64.so
不传参数时遍历 dist/ 下所有 .so。
"""
import ctypes
import glob
import os
import sys

EXPORTS = {
    "singbox": ["StartSingBox", "StopSingBox"],
    "cloudflared": ["StartCloudflared", "StopCloudflared"],
    "nezha": ["StartNezhaAgent", "StopNezhaAgent"],
}


def expected_for(path):
    base = os.path.basename(path)
    for key in EXPORTS:
        if base.startswith(key):
            return EXPORTS[key]
    return []


def check(path):
    lib = ctypes.CDLL(path)
    missing = [n for n in expected_for(path) if not hasattr(lib, n)]
    status = "OK" if not missing else f"MISSING {missing}"
    print(f"{os.path.basename(path):<28} {status}")
    return not missing


def main():
    paths = sys.argv[1:] or sorted(glob.glob("dist/*.so"))
    if not paths:
        print("no .so found; run ./build.sh first")
        return 1
    ok = all(check(p) for p in paths)
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())