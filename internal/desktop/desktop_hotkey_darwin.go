//go:build darwin

package desktop

import "golang.design/x/hotkey"

func shortcutModifier(token string) (hotkey.Modifier, bool) {
	switch token {
	case "option", "alt":
		return hotkey.ModOption, true
	case "ctrl", "control":
		return hotkey.ModCtrl, true
	case "shift":
		return hotkey.ModShift, true
	case "cmd", "command", "meta", "super":
		return hotkey.ModCmd, true
	default:
		return 0, false
	}
}
