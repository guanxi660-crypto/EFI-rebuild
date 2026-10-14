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
	nzMu      sync.Mutex
	nzRunning bool
)

var ShutdownChan = make(chan struct{})

// StartStatus 启动监控 Agent。
// payload 格式: {"config": "/path/to/config.yaml"}
// 返回值：0 成功，-1 已在运行，-2 payload 解析失败。
//
//export StartStatus
func StartStatus(payload *C.char) C.int {
	nzMu.Lock()
	defer nzMu.Unlock()

	if nzRunning {
		return -1
	}

	var p struct {
		Config string `json:"config"`
	}
	if err := json.Unmarshal([]byte(C.GoString(payload)), &p); err != nil {
		return -2
	}

	if p.Config != "" {
		os.Args = []string{"nezha-agent", "-c", p.Config}
	} else {
		os.Args = []string{"nezha-agent"}
	}

	ShutdownChan = make(chan struct{})
	nzRunning = true
	go StatusMain()
	return 0
}

// StopStatus 通知 Agent 退出主循环。
// 返回值：0 成功，-1 当前未运行。
//
//export StopStatus
func StopStatus() C.int {
	nzMu.Lock()
	defer nzMu.Unlock()

	if !nzRunning {
		return -1
	}

	close(ShutdownChan)
	nzRunning = false
	return 0
}

func main() {}