//go:build darwin

package desktop

import (
	"sync"

	"github.com/ebitengine/purego"
)

type cgPoint struct {
	X float64
	Y float64
}

var darwinCursor struct {
	once        sync.Once
	ok          bool
	create      func(uintptr) uintptr
	location    func(uintptr) cgPoint
	release     func(uintptr)
	loadErr     error
	symbolError error
}

func cursorPosition() (int, int, bool) {
	darwinCursor.once.Do(loadDarwinCursorAPI)
	if !darwinCursor.ok {
		return 0, 0, false
	}
	event := darwinCursor.create(0)
	if event == 0 {
		return 0, 0, false
	}
	point := darwinCursor.location(event)
	darwinCursor.release(event)
	return int(point.X), int(point.Y), true
}

func loadDarwinCursorAPI() {
	coreGraphics, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		darwinCursor.loadErr = err
		return
	}
	coreFoundation, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		darwinCursor.loadErr = err
		return
	}
	if !registerDarwinCursorFunc(coreGraphics, "CGEventCreate", &darwinCursor.create) {
		return
	}
	if !registerDarwinCursorFunc(coreGraphics, "CGEventGetLocation", &darwinCursor.location) {
		return
	}
	if !registerDarwinCursorFunc(coreFoundation, "CFRelease", &darwinCursor.release) {
		return
	}
	darwinCursor.ok = true
}

func registerDarwinCursorFunc(handle uintptr, name string, target any) bool {
	symbol, err := purego.Dlsym(handle, name)
	if err != nil {
		darwinCursor.symbolError = err
		return false
	}
	purego.RegisterFunc(target, symbol)
	return true
}
