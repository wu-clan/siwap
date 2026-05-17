package desktop

import "time"

const (
	edgeRevealInterval     = 100 * time.Millisecond
	edgeRevealHoldDuration = time.Second
	edgeHideDelay          = time.Second
	edgeRevealCooldown     = 650 * time.Millisecond
)

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

func (a *App) stopEdgeRevealWatcher() {
	a.edgeMu.Lock()
	defer a.edgeMu.Unlock()
	if a.edgeStop == nil {
		return
	}
	close(a.edgeStop)
	a.edgeStop = nil
}

func (a *App) watchLeftEdgeReveal(stop <-chan struct{}) {
	ticker := time.NewTicker(edgeRevealInterval)
	defer ticker.Stop()
	lastReveal := time.Time{}
	var candidate edgeRevealCandidate
	var outsideSince time.Time
	for {
		select {
		case <-ticker.C:
			if a.shouldHideOutsideMainWindow() {
				if outsideSince.IsZero() {
					outsideSince = time.Now()
					continue
				}
				if time.Since(outsideSince) >= edgeHideDelay {
					_ = a.HideWindow()
					outsideSince = time.Time{}
					candidate = edgeRevealCandidate{}
				}
				continue
			}
			outsideSince = time.Time{}

			if time.Since(lastReveal) < edgeRevealCooldown {
				continue
			}
			x, y, ok := a.leftEdgeCursorPosition()
			if !ok {
				candidate = edgeRevealCandidate{}
				continue
			}
			now := time.Now()
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

func (a *App) shouldHideOutsideMainWindow() bool {
	if a.mainWindow == nil || !a.mainWindow.IsVisible() || !a.config.Preferences().AutohideOnBlur || a.settingsWindowIsVisible() {
		return false
	}
	x, y, ok := cursorPosition()
	if !ok {
		return false
	}
	left, top, width, height, ok := a.mainWindowBounds()
	return ok && !pointInRect(x, y, left, top, width, height)
}

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

func cursorAtExactLeftEdge(x int, y int, left int, top int, height int, hasScreen bool) bool {
	if !hasScreen {
		return x <= 0
	}
	if y < top || y > top+height {
		return false
	}
	return x == left
}

func (a *App) mainWindowBounds() (left int, top int, width int, height int, ok bool) {
	if a.mainWindow == nil {
		return 0, 0, 0, 0, false
	}
	left, top = a.mainWindow.Position()
	width, height = a.mainWindow.Size()
	return left, top, width, height, width > 0 && height > 0
}

func pointInRect(x int, y int, left int, top int, width int, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}
	return x >= left && x < left+width && y >= top && y < top+height
}

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

func (a *App) settingsWindowIsVisible() bool {
	if a.settingsWindow == nil {
		return false
	}
	return a.settingsWindow.IsVisible()
}
