//go:build linux

package desktop

import "golang.design/x/hotkey"

// shortcutModifier 将快捷键修饰符转换为 Linux 对应值
func shortcutModifier(token string) (hotkey.Modifier, bool) {
	switch token {
	case "option", "alt":
		return hotkey.Mod1, true
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "shift":
		return hotkey.ModShift, true
	case "cmd", "command", "meta", "super", "win", "windows":
		return hotkey.Mod4, true
	default:
		return 0, false
	}
}
