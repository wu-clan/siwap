//go:build linux

package desktop

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var linuxCursorTool struct {
	once sync.Once
	path string
}

func cursorPosition() (int, int, bool) {
	linuxCursorTool.once.Do(func() {
		linuxCursorTool.path, _ = exec.LookPath("xdotool")
	})
	if linuxCursorTool.path == "" {
		return 0, 0, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	output, err := exec.CommandContext(ctx, linuxCursorTool.path, "getmouselocation", "--shell").Output()
	if err != nil {
		return 0, 0, false
	}
	return parseXDoToolPosition(string(output))
}

func parseXDoToolPosition(output string) (int, int, bool) {
	var x, y int
	var hasX, hasY bool
	for _, line := range strings.Split(output, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if !found {
			continue
		}
		number, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		switch key {
		case "X":
			x, hasX = number, true
		case "Y":
			y, hasY = number, true
		}
	}
	return x, y, hasX && hasY
}
