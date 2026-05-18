package desktop

import (
	"strings"

	"golang.design/x/hotkey"

	"siwap/internal/domain"
)

// registerShortcut 注册全局呼出快捷键
func (a *App) registerShortcut(shortcut string) {
	mods, key, ok := parseShortcut(shortcut)
	if !ok {
		a.unregisterShortcut()
		return
	}
	a.hotkeyMu.Lock()
	defer a.hotkeyMu.Unlock()
	if a.hotkey != nil && a.hotkeyShortcut == shortcut {
		return
	}
	a.unregisterShortcutLocked()
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		a.emit("shortcut:updated", domain.ActionResult{OK: false, Status: "failed", Message: err.Error()})
		return
	}
	stop := make(chan struct{})
	a.hotkey = hk
	a.hotkeyStop = stop
	a.hotkeyShortcut = shortcut
	a.emit("shortcut:updated", domain.ActionResult{OK: true, Status: "registered", Message: shortcut})
	go func() {
		for {
			select {
			case <-hk.Keydown():
				_ = a.ShowWindow()
			case <-stop:
				return
			}
		}
	}()
}

// unregisterShortcut 注销当前全局快捷键
func (a *App) unregisterShortcut() {
	a.hotkeyMu.Lock()
	defer a.hotkeyMu.Unlock()
	a.unregisterShortcutLocked()
}

// unregisterShortcutLocked 在持锁状态下注销全局快捷键
func (a *App) unregisterShortcutLocked() {
	if a.hotkeyStop != nil {
		close(a.hotkeyStop)
		a.hotkeyStop = nil
	}
	if a.hotkey != nil {
		_ = a.hotkey.Unregister()
		a.hotkey = nil
	}
	a.hotkeyShortcut = ""
}

// parseShortcut 将配置字符串解析为快捷键组合
func parseShortcut(shortcut string) ([]hotkey.Modifier, hotkey.Key, bool) {
	parts := strings.FieldsFunc(shortcut, func(r rune) bool {
		return r == '+' || r == ' ' || r == '-'
	})
	var mods []hotkey.Modifier
	var key hotkey.Key
	hasKey := false
	for _, part := range parts {
		token := strings.ToLower(strings.TrimSpace(part))
		if token == "" {
			continue
		}
		if mod, ok := shortcutModifier(token); ok {
			mods = append(mods, mod)
			continue
		}
		parsed, ok := shortcutKey(token)
		if !ok || hasKey {
			return nil, 0, false
		}
		key = parsed
		hasKey = true
	}
	return mods, key, hasKey
}

// shortcutKey 将按键名称转换为 hotkey 键值
func shortcutKey(token string) (hotkey.Key, bool) {
	keys := map[string]hotkey.Key{
		"a": hotkey.KeyA, "b": hotkey.KeyB, "c": hotkey.KeyC, "d": hotkey.KeyD,
		"e": hotkey.KeyE, "f": hotkey.KeyF, "g": hotkey.KeyG, "h": hotkey.KeyH,
		"i": hotkey.KeyI, "j": hotkey.KeyJ, "k": hotkey.KeyK, "l": hotkey.KeyL,
		"m": hotkey.KeyM, "n": hotkey.KeyN, "o": hotkey.KeyO, "p": hotkey.KeyP,
		"q": hotkey.KeyQ, "r": hotkey.KeyR, "s": hotkey.KeyS, "t": hotkey.KeyT,
		"u": hotkey.KeyU, "v": hotkey.KeyV, "w": hotkey.KeyW, "x": hotkey.KeyX,
		"y": hotkey.KeyY, "z": hotkey.KeyZ, "0": hotkey.Key0, "1": hotkey.Key1,
		"2": hotkey.Key2, "3": hotkey.Key3, "4": hotkey.Key4, "5": hotkey.Key5,
		"6": hotkey.Key6, "7": hotkey.Key7, "8": hotkey.Key8, "9": hotkey.Key9,
		"space": hotkey.KeySpace, "f1": hotkey.KeyF1, "f2": hotkey.KeyF2,
		"f3": hotkey.KeyF3, "f4": hotkey.KeyF4, "f5": hotkey.KeyF5,
		"f6": hotkey.KeyF6, "f7": hotkey.KeyF7, "f8": hotkey.KeyF8,
		"f9": hotkey.KeyF9, "f10": hotkey.KeyF10, "f11": hotkey.KeyF11,
		"f12": hotkey.KeyF12,
	}
	key, ok := keys[token]
	return key, ok
}
