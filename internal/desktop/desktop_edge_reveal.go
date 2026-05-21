package desktop

import "time"

const (
	edgeRevealInterval     = 100 * time.Millisecond
	edgeRevealHoldDuration = time.Second
	edgeRevealCooldown     = 650 * time.Millisecond
)

// startEdgeRevealWatcher 启动左侧边缘呼出监听
func (a *App) startEdgeRevealWatcher() {
	a.edgeMu.Lock()
	if a.edgeStop != nil {
		a.edgeMu.Unlock()
		return
	}
	stop := make(chan struct{})
	a.edgeStop = stop
	a.edgeMu.Unlock()

	go a.watchLeftEdgeReveal(stop)
}

// stopEdgeRevealWatcher 停止左侧边缘呼出监听
func (a *App) stopEdgeRevealWatcher() {
	a.edgeMu.Lock()
	defer a.edgeMu.Unlock()
	if a.edgeStop == nil {
		return
	}
	close(a.edgeStop)
	a.edgeStop = nil
}

// watchLeftEdgeReveal 周期检测鼠标位置并处理边缘呼出
func (a *App) watchLeftEdgeReveal(stop <-chan struct{}) {
	ticker := time.NewTicker(edgeRevealInterval)
	defer ticker.Stop()
	lastReveal := time.Time{}
	var candidate edgeRevealCandidate
	for {
		select {
		case <-ticker.C:
			if time.Since(lastReveal) < edgeRevealCooldown {
				continue
			}
			x, y, ok := a.leftEdgeCursorPosition()
			if !ok {
				candidate = edgeRevealCandidate{}
				continue
			}
			now := time.Now()
			// 鼠标需要在同一个边缘点停留一段时间，避免快速划过屏幕边缘时误触发
			if candidate.since.IsZero() || candidate.x != x || candidate.y != y {
				candidate = edgeRevealCandidate{x: x, y: y, since: now}
				continue
			}
			if now.Sub(candidate.since) < edgeRevealHoldDuration {
				continue
			}
			_ = a.ShowWindow()
			lastReveal = now
			candidate = edgeRevealCandidate{}
		case <-stop:
			return
		}
	}
}

type edgeRevealCandidate struct {
	x     int
	y     int
	since time.Time
}

// leftEdgeCursorPosition 返回位于屏幕左边缘的鼠标坐标
func (a *App) leftEdgeCursorPosition() (int, int, bool) {
	if a.mainWindow == nil || a.mainWindow.IsVisible() {
		return 0, 0, false
	}
	x, y, ok := cursorPosition()
	if !ok {
		return 0, 0, false
	}
	left, top, height, ok := a.primaryScreenEdge()
	if !ok {
		return x, y, cursorAtExactLeftEdge(x, y, 0, 0, 0, false)
	}
	return x, y, cursorAtExactLeftEdge(x, y, left, top, height, true)
}

// cursorAtExactLeftEdge 判断鼠标是否位于屏幕精确左边缘
func cursorAtExactLeftEdge(x int, y int, left int, top int, height int, hasScreen bool) bool {
	if !hasScreen {
		return x <= 0
	}
	if y < top || y > top+height {
		return false
	}
	return x == left
}

// primaryScreenEdge 返回主屏幕左边缘位置
func (a *App) primaryScreenEdge() (left int, top int, height int, ok bool) {
	if a.desktop == nil {
		return 0, 0, 0, false
	}
	screen := a.desktop.Screen.GetPrimary()
	if screen == nil {
		screens := a.desktop.Screen.GetAll()
		if len(screens) == 0 {
			return 0, 0, 0, false
		}
		screen = screens[0]
		for _, item := range screens {
			if item.IsPrimary {
				screen = item
				break
			}
		}
	}
	height = screen.Size.Height
	if height == 0 {
		height = screen.Bounds.Height
	}
	if height == 0 {
		height = screen.WorkArea.Height
	}
	return screen.X, screen.Y, height, height > 0
}

// settingsWindowIsVisible 判断设置窗口是否可见
func (a *App) settingsWindowIsVisible() bool {
	if a.settingsWindow == nil {
		return false
	}
	return a.settingsWindow.IsVisible()
}
