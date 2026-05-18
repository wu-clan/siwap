package desktop

import "siwap/internal/domain"

// currentAdapters 返回已按偏好过滤和排序的终端适配器
func (a *App) currentAdapters() []domain.TerminalAdapter {
	prefs := a.config.Preferences()
	disabled := stringSet(prefs.DisabledTerminalIDs)
	adapters := a.terminals.ListWithProfiles(a.config.ListTerminalProfiles())
	for i := range adapters {
		if adapters[i].ID == "auto" {
			adapters[i].Enabled = true
			continue
		}
		if !adapters[i].Installed {
			adapters[i].Enabled = false
			continue
		}
		if disabled[adapters[i].ID] {
			adapters[i].Enabled = false
		}
	}
	if len(prefs.TerminalOrder) == 0 {
		return adapters
	}
	byID := map[string]domain.TerminalAdapter{}
	for _, adapter := range adapters {
		byID[adapter.ID] = adapter
	}
	ordered := make([]domain.TerminalAdapter, 0, len(adapters))
	seen := map[string]bool{}
	for _, id := range prefs.TerminalOrder {
		if adapter, ok := byID[id]; ok {
			ordered = append(ordered, adapter)
			seen[id] = true
		}
	}
	for _, adapter := range adapters {
		if !seen[adapter.ID] {
			ordered = append(ordered, adapter)
		}
	}
	return ordered
}

// stringSet 将字符串列表转换为集合
func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		if value != "" {
			out[value] = true
		}
	}
	return out
}
