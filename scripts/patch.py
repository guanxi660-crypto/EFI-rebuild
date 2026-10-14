#!/usr/bin/env python3
"""在上游源码中做最小改动，使其可作为 package main 编译为 c-shared 动态库。

cloudflared / nezha-agent 的入口都在 `package main` 中，无法直接被外部导入，
因此构建时对 clone 下来的源码做少量文本替换：
  * 将 `func main()` 重命名为可调用的导出入口；
  * 暴露优雅关闭所需的通道；
  * 移除会直接终止宿主进程的 log.Fatal。

xray-core 无需补丁（其核心逻辑位于可导入的子包中）。
"""
import sys
import os


def _read(path):
    with open(path, "r", encoding="utf-8") as f:
        return f.read()


def _write(path, content):
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(content)


def _must_replace(content, old, new, path):
    if old not in content:
        raise SystemExit(f"patch failed: pattern not found in {path}:\n{old}")
    return content.replace(old, new)


def patch_cloudflared(src):
    path = os.path.join(src, "cmd", "cloudflared", "main.go")
    content = _read(path)

    content = _must_replace(content, "func main() {", "func CloudflaredMain() {", path)

    # 让 main 使用 export.go 中定义的包级优雅关闭通道，便于 StopCloudflared 触发关闭。
    content = _must_replace(
        content,
        "graceShutdownC := make(chan struct{})",
        "graceShutdownC := GraceShutdownC",
        path,
    )

    _write(path, content)


def patch_nezha(src):
    path = os.path.join(src, "cmd", "agent", "main.go")
    content = _read(path)

    content = _must_replace(content, "func main() {", "func NezhaMain() {", path)

    # 在主循环 select 中增加关闭分支，使 StopNezha 能优雅退出。
    content = _must_replace(
        content,
        "case <-reloadSigChan:",
        "case <-ShutdownChan:\n\t\t\tprintln(\"Shutdown signal received...\")\n\t\t\treturn\n\t\tcase <-reloadSigChan:",
        path,
    )

    # 避免 log.Fatal 直接 os.Exit 终止宿主进程（改用 log.Printf 保留 log 依赖）。
    content = content.replace("log.Fatal(err)", "log.Printf(\"agent exited: %v\", err)")
    content = content.replace("log.Fatalf(", "log.Printf(")

    _write(path, content)


def main():
    if len(sys.argv) != 3:
        raise SystemExit("usage: patch.py <cloudflared|nezha> <src-dir>")
    target, src = sys.argv[1], sys.argv[2]
    if target == "cloudflared":
        patch_cloudflared(src)
    elif target == "nezha":
        patch_nezha(src)
    else:
        raise SystemExit(f"unknown target: {target}")
    print(f"patched {target} at {src}")


if __name__ == "__main__":
    main()