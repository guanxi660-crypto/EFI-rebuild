#!/usr/bin/env python3
"""sing-box 生命周期测试：通过 FFI 启动一个 SOCKS 入站再关闭。

用法：
    python3 tests/ffi_singbox_start.py dist/singbox_amd64.so
"""
import ctypes
import json
import os
import sys
import tempfile
import time


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else "dist/singbox_amd64.so"
    lib = ctypes.CDLL(path)

    lib.StartSingBox.argtypes = [ctypes.c_char_p]
    lib.StartSingBox.restype = ctypes.c_int
    lib.StopSingBox.restype = ctypes.c_int

    config = {
        "log": {"level": "warn"},
        "inbounds": [
            {
                "type": "socks",
                "tag": "socks-in",
                "listen": "127.0.0.1",
                "listen_port": 10808,
            }
        ],
        "outbounds": [{"type": "direct", "tag": "direct"}],
    }

    with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
        json.dump(config, f)
        config_path = f.name

    payload = json.dumps({"config": config_path, "workingDir": ".", "disableColor": True})

    try:
        rc = lib.StartSingBox(payload.encode("utf-8"))
        print("StartSingBox ->", rc)
        if rc != 0:
            return 1

        time.sleep(1.5)
        print("StopSingBox ->", lib.StopSingBox())
        return 0
    finally:
        os.unlink(config_path)


if __name__ == "__main__":
    sys.exit(main())