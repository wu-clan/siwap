package config

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"siwap/internal/domain"
)

const currentVersion = 3
const AllProjectsScopeID = "__all_projects"

type Store struct {
	mu      sync.Mutex
	path    string
	config  domain.AppConfig
	Summary domain.AppSummary
}

func NewStore() *Store {
	path := filepath.Join(appDataDir(), "config.json")
	store := &Store{
		path:    path,
		config:  defaultConfig(),
		Summary: defaultSummary(),
	}
	_ = store.Load()
	return store
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.saveLocked()
		}
		return err
	}

	loaded := defaultConfig()
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}
	s.config = mergeConfig(loaded)
	return s.saveLocked()
}

func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) ConfigPath() string {
	return s.path
}

func (s *Store) Preferences() domain.Preferences {
	s.mu.Lock()
	defer s.mu.Unlock()
	return clonePreferences(s.config.Preferences)
}

func (s *Store) UpdatePreferences(next domain.Preferences) (domain.Preferences, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefs := s.config.Preferences
	prefs.SelectedProjectID = next.SelectedProjectID
	prefs.DefaultProjectID = next.DefaultProjectID
	prefs.Language = firstNonEmpty(next.Language, prefs.Language)
	prefs.Appearance = firstNonEmpty(next.Appearance, prefs.Appearance)
	if next.DefaultAdapterID != "" {
		prefs.DefaultAdapterID = next.DefaultAdapterID
	}
	if next.TerminalCommandTemplate != "" {
		prefs.TerminalCommandTemplate = next.TerminalCommandTemplate
	}
	if next.TerminalOrder != nil {
		prefs.TerminalOrder = append([]string(nil), next.TerminalOrder...)
	}
	if next.DisabledTerminalIDs != nil {
		prefs.DisabledTerminalIDs = append([]string(nil), next.DisabledTerminalIDs...)
	}
	if next.HarnessOrder != nil {
		prefs.HarnessOrder = append([]string(nil), next.HarnessOrder...)
	}
	prefs.GlobalShortcut = firstNonEmpty(next.GlobalShortcut, prefs.GlobalShortcut)
	prefs.LaunchInBackground = next.LaunchInBackground
	prefs.WorktreeBaseDir = next.WorktreeBaseDir
	prefs.WorktreeLocation = firstNonEmpty(next.WorktreeLocation, prefs.WorktreeLocation)
	prefs.AutohideOnBlur = next.AutohideOnBlur
	prefs.PanelWidth = positiveOr(next.PanelWidth, prefs.PanelWidth)
	prefs.WindowWidth = positiveOr(next.WindowWidth, prefs.WindowWidth)
	prefs.WindowHeight = positiveOr(next.WindowHeight, prefs.WindowHeight)
	prefs.WindowX = next.WindowX
	prefs.WindowY = next.WindowY
	prefs.WindowPositionSaved = next.WindowPositionSaved
	prefs.AlwaysOnTop = next.AlwaysOnTop
	s.config.Preferences = prefs
	return clonePreferences(prefs), s.saveLocked()
}

func (s *Store) ListProjects() []domain.Project {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneProjects(s.config.Projects)
}

func (s *Store) GetProject(id string) (domain.Project, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, project := range s.config.Projects {
		if project.ID == id {
			return project, true
		}
	}
	return domain.Project{}, false
}

func (s *Store) SelectedProject() (domain.Project, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	selectedID := s.config.Preferences.SelectedProjectID
	if selectedID == "" {
		return domain.Project{}, false
	}
	for _, project := range s.config.Projects {
		if project.ID == selectedID {
			return project, true
		}
	}
	return domain.Project{}, false
}

func (s *Store) AddProject(path string, label string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clean, err := normalizePath(path)
	if err != nil {
		return domain.Project{}, err
	}
	info, err := os.Stat(clean)
	if err != nil {
		return domain.Project{}, err
	}
	if !info.IsDir() {
		return domain.Project{}, fmt.Errorf("project path is not a directory: %s", clean)
	}

	for i, project := range s.config.Projects {
		if samePath(project.Path, clean) {
			s.config.Projects[i].Label = fallbackLabel(clean, "")
			if s.config.Preferences.SelectedProjectID == "" {
				s.config.Preferences.SelectedProjectID = project.ID
			}
			return s.config.Projects[i], s.saveLocked()
		}
	}

	isFirst := len(s.config.Projects) == 0
	project := domain.Project{
		ID:         projectID(clean),
		Path:       clean,
		Label:      fallbackLabel(clean, ""),
		IsDefault:  isFirst,
		LastUsedAt: now(),
	}
	s.config.Projects = append(s.config.Projects, project)
	if isFirst || s.config.Preferences.DefaultProjectID == "" {
		s.config.Preferences.DefaultProjectID = project.ID
	}
	if isFirst || s.config.Preferences.SelectedProjectID == "" {
		s.config.Preferences.SelectedProjectID = project.ID
	}
	return project, s.saveLocked()
}

func (s *Store) RemoveProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	projects := s.config.Projects[:0]
	removed := false
	for _, project := range s.config.Projects {
		if project.ID == id {
			removed = true
			continue
		}
		projects = append(projects, project)
	}
	if !removed {
		return fmt.Errorf("project not found: %s", id)
	}
	s.config.Projects = projects
	if s.config.Preferences.SelectedProjectID == id {
		s.config.Preferences.SelectedProjectID = ""
	}
	if s.config.Preferences.DefaultProjectID == id {
		s.config.Preferences.DefaultProjectID = ""
	}
	s.ensureProjectDefaultsLocked()
	return s.saveLocked()
}

func (s *Store) SelectProject(id string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" || id == AllProjectsScopeID {
		s.config.Preferences.SelectedProjectID = AllProjectsScopeID
		return domain.Project{}, s.saveLocked()
	}
	for i, project := range s.config.Projects {
		if project.ID == id {
			s.config.Projects[i].LastUsedAt = now()
			s.config.Preferences.SelectedProjectID = id
			return s.config.Projects[i], s.saveLocked()
		}
	}
	return domain.Project{}, fmt.Errorf("project not found: %s", id)
}

func (s *Store) SetDefaultProject(id string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, project := range s.config.Projects {
		isDefault := project.ID == id
		s.config.Projects[i].IsDefault = isDefault
		if isDefault {
			s.config.Preferences.DefaultProjectID = id
			s.config.Preferences.SelectedProjectID = id
		}
	}
	if s.config.Preferences.DefaultProjectID != id {
		return domain.Project{}, fmt.Errorf("project not found: %s", id)
	}
	for _, project := range s.config.Projects {
		if project.ID == id {
			return project, s.saveLocked()
		}
	}
	return domain.Project{}, fmt.Errorf("project not found: %s", id)
}

func (s *Store) ReorderProjects(ids []string) ([]domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(ids) == 0 {
		return cloneProjects(s.config.Projects), nil
	}
	byID := map[string]domain.Project{}
	for _, project := range s.config.Projects {
		byID[project.ID] = project
	}
	next := make([]domain.Project, 0, len(s.config.Projects))
	seen := map[string]bool{}
	for _, id := range ids {
		project, ok := byID[id]
		if !ok || seen[id] {
			continue
		}
		next = append(next, project)
		seen[id] = true
	}
	for _, project := range s.config.Projects {
		if !seen[project.ID] {
			next = append(next, project)
		}
	}
	s.config.Projects = next
	s.ensureProjectDefaultsLocked()
	return cloneProjects(s.config.Projects), s.saveLocked()
}

func (s *Store) ListHarnesses() []domain.Harness {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneHarnesses(s.config.Harnesses)
}

func (s *Store) GetHarness(id string) (domain.Harness, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, harness := range s.config.Harnesses {
		if harness.ID == id {
			return harness, true
		}
	}
	return domain.Harness{}, false
}

func (s *Store) UpdateHarness(next domain.Harness) (domain.Harness, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, harness := range s.config.Harnesses {
		if harness.ID == next.ID {
			if next.Label != "" {
				harness.Label = next.Label
			}
			if next.Command != "" {
				harness.Command = strings.TrimSpace(next.Command)
			}
			harness.Enabled = next.Enabled
			if next.Icon != "" {
				harness.Icon = next.Icon
			}
			if next.IconSource != "" {
				harness.IconSource = next.IconSource
			}
			if next.Tint != "" {
				harness.Tint = next.Tint
			}
			harness.BuiltIn = harness.BuiltIn || next.BuiltIn
			if next.Flags != nil {
				harness.Flags = cloneStringMap(next.Flags)
			}
			s.config.Harnesses[i] = harness
			return harness, s.saveLocked()
		}
	}
	return domain.Harness{}, fmt.Errorf("harness not found: %s", next.ID)
}

func (s *Store) CreateHarness(next domain.Harness) (domain.Harness, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next.ID = strings.TrimSpace(next.ID)
	if next.ID == "" {
		next.ID = digestID("harness", next.Label+next.Command+time.Now().String())
	}
	next.Label = strings.TrimSpace(next.Label)
	next.Command = strings.TrimSpace(next.Command)
	if next.Label == "" || next.Command == "" {
		return domain.Harness{}, errors.New("assistant name and command are required")
	}
	if next.Icon == "" {
		next.Icon = "custom"
	}
	if next.IconSource == "" {
		next.IconSource = "custom"
	}
	if next.Tint == "" {
		next.Tint = "#5E9CFF"
	}
	if next.Flags == nil {
		next.Flags = map[string]string{}
	}
	next.Enabled = true
	for _, item := range s.config.Harnesses {
		if item.ID == next.ID {
			return domain.Harness{}, fmt.Errorf("assistant already exists: %s", next.ID)
		}
	}
	s.config.Harnesses = append(s.config.Harnesses, next)
	return next, s.saveLocked()
}

func (s *Store) RemoveHarness(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	harnesses := s.config.Harnesses[:0]
	removed := false
	for _, item := range s.config.Harnesses {
		if item.ID == id {
			if item.BuiltIn {
				return fmt.Errorf("built-in assistant cannot be removed: %s", id)
			}
			removed = true
			continue
		}
		harnesses = append(harnesses, item)
	}
	if !removed {
		return fmt.Errorf("assistant not found: %s", id)
	}
	s.config.Harnesses = harnesses
	s.config.Preferences.HarnessOrder = removeString(s.config.Preferences.HarnessOrder, id)
	return s.saveLocked()
}

func (s *Store) ReorderHarnesses(ids []string) ([]domain.Harness, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(ids) == 0 {
		return cloneHarnesses(s.config.Harnesses), nil
	}
	byID := map[string]domain.Harness{}
	for _, item := range s.config.Harnesses {
		byID[item.ID] = item
	}
	next := make([]domain.Harness, 0, len(s.config.Harnesses))
	seen := map[string]bool{}
	for _, id := range ids {
		item, ok := byID[id]
		if !ok || seen[id] {
			continue
		}
		next = append(next, item)
		seen[id] = true
	}
	for _, item := range s.config.Harnesses {
		if !seen[item.ID] {
			next = append(next, item)
		}
	}
	s.config.Harnesses = next
	s.config.Preferences.HarnessOrder = ids
	return cloneHarnesses(s.config.Harnesses), s.saveLocked()
}

func (s *Store) ListTerminalProfiles() []domain.TerminalProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneProfiles(s.config.TerminalProfiles)
}

func (s *Store) UpsertTerminalProfile(profile domain.TerminalProfile) (domain.TerminalProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile.Label = strings.TrimSpace(profile.Label)
	profile.ExecutablePath = strings.TrimSpace(profile.ExecutablePath)
	if profile.ID == "" {
		profile.ID = profileID(profile.Label + profile.ExecutablePath + time.Now().String())
	}
	if profile.Label == "" || profile.ExecutablePath == "" {
		return domain.TerminalProfile{}, errors.New("profile label and executable path are required")
	}
	if profile.ArgumentTemplate == "" {
		profile.ArgumentTemplate = "{{command}}"
	}
	if profile.CommandMode == "" {
		profile.CommandMode = "shell"
	}
	if profile.Platform == "" {
		profile.Platform = runtime.GOOS
	}
	profile.Enabled = true
	for i, item := range s.config.TerminalProfiles {
		if item.ID == profile.ID {
			s.config.TerminalProfiles[i] = profile
			return profile, s.saveLocked()
		}
	}
	s.config.TerminalProfiles = append(s.config.TerminalProfiles, profile)
	return profile, s.saveLocked()
}

func (s *Store) RemoveTerminalProfile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles := s.config.TerminalProfiles[:0]
	removed := false
	for _, profile := range s.config.TerminalProfiles {
		if profile.ID == id {
			removed = true
			continue
		}
		profiles = append(profiles, profile)
	}
	if !removed {
		return fmt.Errorf("terminal profile not found: %s", id)
	}
	s.config.TerminalProfiles = profiles
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) ensureProjectDefaultsLocked() {
	normalizeProjects(&s.config)
}

func normalizeProjects(config *domain.AppConfig) {
	if len(config.Projects) == 0 {
		config.Preferences.SelectedProjectID = AllProjectsScopeID
		config.Preferences.DefaultProjectID = ""
		return
	}
	defaultID := config.Preferences.DefaultProjectID
	if defaultID == "" {
		for _, project := range config.Projects {
			if project.IsDefault {
				defaultID = project.ID
				break
			}
		}
	}
	if defaultID == "" {
		defaultID = config.Projects[0].ID
	}
	defaultExists := false
	selectedExists := false
	for i := range config.Projects {
		config.Projects[i].Label = fallbackLabel(config.Projects[i].Path, "")
		isDefault := config.Projects[i].ID == defaultID
		config.Projects[i].IsDefault = isDefault
		if isDefault {
			defaultExists = true
		}
		if config.Projects[i].ID == config.Preferences.SelectedProjectID {
			selectedExists = true
		}
	}
	if !defaultExists {
		defaultID = config.Projects[0].ID
		config.Projects[0].IsDefault = true
	}
	config.Preferences.DefaultProjectID = defaultID
	if config.Preferences.SelectedProjectID == "" || !selectedExists {
		config.Preferences.SelectedProjectID = defaultID
	}
}

func defaultSummary() domain.AppSummary {
	return domain.AppSummary{
		Name: "Siwap",
		Stack: []string{
			"Wails v3",
			"Go backend",
			"Vue 3 + TypeScript + Vite Plus frontend",
			"pnpm package management",
			"Terminal adapter architecture",
		},
		Scope: []string{
			"Cross-platform desktop shell",
			"Multi-terminal session launcher",
			"Project-aware harness execution",
			"Worktree orchestration",
		},
		Exclusions: []string{},
	}
}

func defaultConfig() domain.AppConfig {
	return domain.AppConfig{
		Version:          currentVersion,
		Harnesses:        defaultHarnesses(),
		Projects:         []domain.Project{},
		TerminalProfiles: []domain.TerminalProfile{},
		Preferences: domain.Preferences{
			SelectedProjectID:       AllProjectsScopeID,
			DefaultProjectID:        "",
			Language:                "zh-CN",
			Appearance:              "system",
			DefaultAdapterID:        "auto",
			TerminalCommandTemplate: "{{command}}",
			TerminalOrder:           []string{},
			DisabledTerminalIDs:     []string{},
			HarnessOrder:            []string{},
			GlobalShortcut:          defaultShortcut(),
			LaunchInBackground:      false,
			WorktreeBaseDir:         "",
			WorktreeLocation:        "project-parent",
			AutohideOnBlur:          true,
			PanelWidth:              0,
			WindowWidth:             320,
			WindowHeight:            900,
			AlwaysOnTop:             false,
		},
	}
}

func defaultHarnesses() []domain.Harness {
	return []domain.Harness{
		{
			ID:         "claude-code",
			Label:      "Claude Code",
			Command:    "claude",
			Enabled:    true,
			BuiltIn:    true,
			Icon:       "claude",
			IconSource: "builtin",
			Tint:       "#F2A65A",
			Flags:      map[string]string{"dangerouslySkipPermissions": "false", "permissionMode": "default"},
			FlagOptions: []domain.HarnessFlag{
				{Key: "dangerouslySkipPermissions", Label: "Skip permissions", Type: "toggle", CommandFlag: "--dangerously-skip-permissions", Default: "false"},
				{Key: "permissionMode", Label: "Permission mode", Type: "select", CommandFlag: "--permission-mode", Default: "default", Options: []string{"default", "acceptEdits", "plan", "auto", "dontAsk", "bypassPermissions"}},
			},
		},
		{
			ID:         "codex",
			Label:      "Codex",
			Command:    "codex",
			Enabled:    true,
			BuiltIn:    true,
			Icon:       "codex",
			IconSource: "builtin",
			Tint:       "#66D9E8",
			Flags:      map[string]string{"yolo": "false"},
			FlagOptions: []domain.HarnessFlag{
				{Key: "yolo", Label: "YOLO mode", Type: "toggle", CommandFlag: "--yolo", Default: "false"},
			},
		},
		{
			ID:         "opencode",
			Label:      "OpenCode",
			Command:    "opencode",
			Enabled:    true,
			BuiltIn:    true,
			Icon:       "opencode",
			IconSource: "builtin",
			Tint:       "#8CE99A",
			Flags:      map[string]string{"yolo": "false"},
			FlagOptions: []domain.HarnessFlag{
				{Key: "yolo", Label: "YOLO mode", Type: "toggle", CommandFlag: "--yolo", Default: "false"},
			},
		},
	}
}

func mergeConfig(loaded domain.AppConfig) domain.AppConfig {
	base := defaultConfig()
	loadedVersion := loaded.Version
	if loaded.Version > 0 {
		base.Version = loaded.Version
	}
	if len(loaded.Harnesses) > 0 {
		base.Harnesses = mergeHarnesses(base.Harnesses, loaded.Harnesses)
	}
	if len(loaded.Projects) > 0 {
		base.Projects = loaded.Projects
		normalizeProjects(&base)
	}
	base.TerminalProfiles = cloneProfiles(loaded.TerminalProfiles)
	base.Preferences = mergePreferences(base.Preferences, loaded.Preferences)
	if loadedVersion < 2 {
		base.Preferences.AutohideOnBlur = true
	}
	if loadedVersion < 3 && isLegacyDefaultShortcut(loaded.Preferences.GlobalShortcut) {
		base.Preferences.GlobalShortcut = defaultShortcut()
	}
	base.Version = currentVersion
	normalizeProjects(&base)
	return base
}

func mergePreferences(base domain.Preferences, loaded domain.Preferences) domain.Preferences {
	if loaded.SelectedProjectID != "" || loaded.SelectedProjectID == "" {
		base.SelectedProjectID = loaded.SelectedProjectID
	}
	base.DefaultProjectID = loaded.DefaultProjectID
	base.Language = firstNonEmpty(loaded.Language, base.Language)
	base.Appearance = firstNonEmpty(loaded.Appearance, base.Appearance)
	base.DefaultAdapterID = firstNonEmpty(loaded.DefaultAdapterID, base.DefaultAdapterID)
	base.TerminalCommandTemplate = firstNonEmpty(loaded.TerminalCommandTemplate, base.TerminalCommandTemplate)
	if loaded.TerminalOrder != nil {
		base.TerminalOrder = append([]string(nil), loaded.TerminalOrder...)
	}
	if loaded.DisabledTerminalIDs != nil {
		base.DisabledTerminalIDs = append([]string(nil), loaded.DisabledTerminalIDs...)
	}
	if loaded.HarnessOrder != nil {
		base.HarnessOrder = append([]string(nil), loaded.HarnessOrder...)
	}
	base.GlobalShortcut = firstNonEmpty(loaded.GlobalShortcut, base.GlobalShortcut)
	base.LaunchInBackground = loaded.LaunchInBackground
	base.WorktreeBaseDir = loaded.WorktreeBaseDir
	base.WorktreeLocation = firstNonEmpty(loaded.WorktreeLocation, base.WorktreeLocation)
	base.AutohideOnBlur = loaded.AutohideOnBlur
	base.PanelWidth = positiveOr(loaded.PanelWidth, base.PanelWidth)
	base.WindowWidth = positiveOr(loaded.WindowWidth, base.WindowWidth)
	base.WindowHeight = positiveOr(loaded.WindowHeight, base.WindowHeight)
	base.WindowX = loaded.WindowX
	base.WindowY = loaded.WindowY
	base.WindowPositionSaved = loaded.WindowPositionSaved
	base.AlwaysOnTop = loaded.AlwaysOnTop
	return base
}

func mergeHarnesses(defaults []domain.Harness, loaded []domain.Harness) []domain.Harness {
	byID := map[string]domain.Harness{}
	for _, harness := range defaults {
		byID[harness.ID] = harness
	}
	out := make([]domain.Harness, 0, len(defaults)+len(loaded))
	seen := map[string]bool{}
	for _, harness := range loaded {
		if current, ok := byID[harness.ID]; ok {
			current.Label = firstNonEmpty(harness.Label, current.Label)
			current.Command = firstNonEmpty(harness.Command, current.Command)
			current.Enabled = harness.Enabled
			current.BuiltIn = true
			current.Icon = firstNonEmpty(harness.Icon, current.Icon)
			current.IconSource = firstNonEmpty(harness.IconSource, current.IconSource)
			current.Tint = firstNonEmpty(harness.Tint, current.Tint)
			if harness.Flags != nil {
				current.Flags = cloneStringMap(harness.Flags)
			}
			byID[harness.ID] = current
			out = append(out, current)
			seen[harness.ID] = true
			continue
		}
		if harness.ID != "" {
			out = append(out, harness)
			seen[harness.ID] = true
		}
	}
	for _, harness := range defaults {
		if !seen[harness.ID] {
			out = append(out, byID[harness.ID])
		}
	}
	return out
}

func cloneConfig(config domain.AppConfig) domain.AppConfig {
	config.Harnesses = cloneHarnesses(config.Harnesses)
	config.Projects = cloneProjects(config.Projects)
	config.TerminalProfiles = cloneProfiles(config.TerminalProfiles)
	config.Preferences = clonePreferences(config.Preferences)
	return config
}

func clonePreferences(in domain.Preferences) domain.Preferences {
	in.TerminalOrder = append([]string(nil), in.TerminalOrder...)
	in.DisabledTerminalIDs = append([]string(nil), in.DisabledTerminalIDs...)
	in.HarnessOrder = append([]string(nil), in.HarnessOrder...)
	return in
}

func cloneHarnesses(in []domain.Harness) []domain.Harness {
	out := make([]domain.Harness, len(in))
	for i, harness := range in {
		out[i] = harness
		out[i].Flags = cloneStringMap(harness.Flags)
		if harness.FlagOptions != nil {
			out[i].FlagOptions = append([]domain.HarnessFlag(nil), harness.FlagOptions...)
		}
	}
	return out
}

func cloneProjects(in []domain.Project) []domain.Project {
	out := make([]domain.Project, len(in))
	copy(out, in)
	return out
}

func cloneProfiles(in []domain.TerminalProfile) []domain.TerminalProfile {
	out := make([]domain.TerminalProfile, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := map[string]string{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func appDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".siwap"
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Siwap")
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "siwap")
	}
	return filepath.Join(home, ".config", "siwap")
}

func normalizePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return filepath.Abs(path)
}

func samePath(a, b string) bool {
	cleanA, errA := normalizePath(a)
	cleanB, errB := normalizePath(b)
	if errA != nil || errB != nil {
		return a == b
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(cleanA, cleanB)
	}
	return cleanA == cleanB
}

func projectID(path string) string {
	return digestID("project", strings.ToLower(path))
}

func profileID(seed string) string {
	return digestID("profile", seed)
}

func digestID(prefix string, seed string) string {
	sum := sha1.Sum([]byte(seed))
	return prefix + "-" + hex.EncodeToString(sum[:])[:12]
}

func fallbackLabel(path string, label string) string {
	if strings.TrimSpace(label) != "" {
		return strings.TrimSpace(label)
	}
	base := filepath.Base(path)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return path
	}
	return base
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func firstNonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func positiveOr(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func removeString(values []string, remove string) []string {
	out := values[:0]
	for _, value := range values {
		if value != remove {
			out = append(out, value)
		}
	}
	return out
}

func defaultShortcut() string {
	return "Control+Command+S"
}

func isLegacyDefaultShortcut(shortcut string) bool {
	switch strings.ToLower(strings.ReplaceAll(shortcut, " ", "")) {
	case "", "option+s", "alt+s":
		return true
	default:
		return false
	}
}
