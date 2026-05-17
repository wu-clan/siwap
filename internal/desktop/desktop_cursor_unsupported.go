//go:build !darwin && !windows && !linux

package desktop

func cursorPosition() (int, int, bool) {
	return 0, 0, false
}
