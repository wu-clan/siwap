//go:build !darwin && !windows && !linux

package desktop

// cursorPosition 在不支持的平台返回不可用状态
func cursorPosition() (int, int, bool) {
	return 0, 0, false
}
