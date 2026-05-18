//go:build windows

package desktop

import "golang.design/x/hotkey"

// shortcutModifier 将快捷键修饰符转换为 Windows 对应值
func shortcutModifier(token string) (hotkey.Modifier, bool) {
	switch token {
	case "option", "alt":
		return hotkey.ModAlt, true
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "shift":
		return hotkey.ModShift, true
	case "cmd", "command", "meta", "super", "win", "windows":
		return hotkey.ModWin, true
	default:
		return 0, false
	}
}
