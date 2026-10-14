package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	cfMu      sync.Mutex
	cfRunning bool
)

var GraceShutdownC = make(chan struct{})

// StartTools 以 args JSON payload 启动隧道。
// payload 格式: {"args": ["tunnel", "--no-autoupdate", "run", "--token", "..."]}
// 返回值：0 成功，-1 已在运行，-2 payload 解析失败。
//
//export StartTools
func StartTools(payload *C.char) C.int {
	cfMu.Lock()
	defer cfMu.Unlock()

	if cfRunning {
		return -1
	}

	var p struct {
		Args []string `json:"args"`
	}
	if err := json.Unmarshal([]byte(C.GoString(payload)), &p); err != nil {
		return -2
	}

	if len(p.Args) == 0 {
		return -2
	}

	GraceShutdownC = make(chan struct{})
	os.Args = append([]string{"cloudflared"}, p.Args...)

	cfRunning = true
	go ToolsMain()
	return 0
}

// StopTools 触发优雅关闭。
// 返回值：0 成功，-1 当前未运行。
//
//export StopTools
func StopTools() C.int {
	cfMu.Lock()
	defer cfMu.Unlock()

	if !cfRunning {
		return -1
	}

	close(GraceShutdownC)
	cfRunning = false
	return 0
}

func main() {}