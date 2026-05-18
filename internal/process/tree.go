package process

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TreePIDs 返回指定进程及其所有子进程的 PID 列表
func TreePIDs(root int) []int {
	if root <= 0 {
		return nil
	}
	// 记录启动器及其子进程，关闭会话时不依赖平台窗口 ID 也能清理进程树
	seen := map[int]bool{root: true}
	queue := []int{root}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, child := range childPIDs(pid) {
			if !seen[child] {
				seen[child] = true
				queue = append(queue, child)
			}
		}
	}
	out := make([]int, 0, len(seen))
	for pid := range seen {
		out = append(out, pid)
	}
	return out
}

// childPIDs 返回指定父进程的直接子进程 PID
func childPIDs(parent int) []int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "wmic", "process", "where", "ParentProcessId="+strconv.Itoa(parent), "get", "ProcessId", "/value")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		return parseWindowsPIDs(string(out))
	}
	cmd := exec.CommandContext(ctx, "pgrep", "-P", strconv.Itoa(parent))
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseLinePIDs(string(out))
}

// parseLinePIDs 从按行输出的进程列表中解析 PID
func parseLinePIDs(data string) []int {
	var out []int
	for _, field := range strings.Fields(data) {
		pid, err := strconv.Atoi(strings.TrimSpace(field))
		if err == nil && pid > 0 {
			out = append(out, pid)
		}
	}
	return out
}

// parseWindowsPIDs 从 Windows wmic 输出中解析 PID
func parseWindowsPIDs(data string) []int {
	var out []int
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ProcessId=") {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimPrefix(line, "ProcessId="))
		if err == nil && pid > 0 {
			out = append(out, pid)
		}
	}
	return out
}
