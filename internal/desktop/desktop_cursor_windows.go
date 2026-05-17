//go:build windows

package desktop

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type winPoint struct {
	X int32
	Y int32
}

var getCursorPos = windows.NewLazySystemDLL("user32.dll").NewProc("GetCursorPos")

func cursorPosition() (int, int, bool) {
	var point winPoint
	result, _, _ := getCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	if result == 0 {
		return 0, 0, false
	}
	return int(point.X), int(point.Y), true
}
