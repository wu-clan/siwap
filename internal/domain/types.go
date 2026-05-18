package domain

// SessionEnvironmentKey 是注入终端环境中的会话 ID 键名
const SessionEnvironmentKey = "SIWAP_SESSION_ID"

// AppSummary 描述应用的基础信息和能力范围
type AppSummary struct {
	Name       string   `json:"name"`
	Stack      []string `json:"stack"`
	Scope      []string `json:"scope"`
	Exclusions []string `json:"exclusions"`
}

// AppConfig 表示持久化到本地的完整应用配置
type AppConfig struct {
	Version          int               `json:"version"`
	Harnesses        []Harness         `json:"harnesses"`
	Projects         []Project         `json:"projects"`
	TerminalProfiles []TerminalProfile `json:"terminalProfiles"`
	Preferences      Preferences       `json:"preferences"`
}

// Preferences 表示用户偏好设置
type Preferences struct {
	SelectedProjectID       string   `json:"selectedProjectId"`
	DefaultProjectID        string   `json:"defaultProjectId"`
	Language                string   `json:"language"`
	Appearance              string   `json:"appearance"`
	DefaultAdapterID        string   `json:"defaultAdapterId"`
	TerminalCommandTemplate string   `json:"terminalCommandTemplate"`
	TerminalOrder           []string `json:"terminalOrder"`
	DisabledTerminalIDs     []string `json:"disabledTerminalIds"`
	HarnessOrder            []string `json:"harnessOrder"`
	GlobalShortcut          string   `json:"globalShortcut"`
	LaunchInBackground      bool     `json:"launchInBackground"`
	WorktreeBaseDir         string   `json:"worktreeBaseDir"`
	WorktreeLocation        string   `json:"worktreeLocation"`
	AutohideOnBlur          bool     `json:"autohideOnBlur"`
	PanelWidth              int      `json:"panelWidth"`
	WindowWidth             int      `json:"windowWidth"`
	WindowHeight            int      `json:"windowHeight"`
	WindowX                 int      `json:"windowX"`
	WindowY                 int      `json:"windowY"`
	WindowPositionSaved     bool     `json:"windowPositionSaved"`
	AlwaysOnTop             bool     `json:"alwaysOnTop"`
}

// TerminalProfile 表示用户自定义终端配置
type TerminalProfile struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	ExecutablePath   string `json:"executablePath"`
	ArgumentTemplate string `json:"argumentTemplate"`
	WorkingDirFlag   string `json:"workingDirFlag"`
	CommandMode      string `json:"commandMode"`
	Platform         string `json:"platform"`
	Enabled          bool   `json:"enabled"`
}

// Harness 表示一个可启动的 AI 助手配置
type Harness struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	Command     string            `json:"command"`
	Enabled     bool              `json:"enabled"`
	BuiltIn     bool              `json:"builtIn"`
	Icon        string            `json:"icon"`
	IconSource  string            `json:"iconSource"`
	Tint        string            `json:"tint"`
	Flags       map[string]string `json:"flags"`
	FlagOptions []HarnessFlag     `json:"flagOptions"`
}

// HarnessFlag 表示助手命令支持的可配置参数
type HarnessFlag struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	CommandFlag string   `json:"commandFlag"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

// Project 表示一个用户项目目录
type Project struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	Label      string `json:"label,omitempty"`
	IsDefault  bool   `json:"isDefault"`
	LastUsedAt string `json:"lastUsedAt,omitempty"`
}

// TerminalCapability 描述终端适配器的一项能力
type TerminalCapability struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Supported   bool   `json:"supported"`
	Description string `json:"description"`
}

// TerminalAdapter 描述一个可用终端适配器
type TerminalAdapter struct {
	ID           string               `json:"id"`
	Label        string               `json:"label"`
	Platform     string               `json:"platform"`
	Executable   string               `json:"executable,omitempty"`
	Installed    bool                 `json:"installed"`
	Enabled      bool                 `json:"enabled"`
	Stability    string               `json:"stability"`
	Confidence   string               `json:"confidence"`
	Message      string               `json:"message,omitempty"`
	Capabilities []TerminalCapability `json:"capabilities"`
}

// TerminalSessionRef 保存终端会话的可追踪引用信息
type TerminalSessionRef struct {
	AdapterID             string   `json:"adapterId"`
	Platform              string   `json:"platform"`
	PID                   int      `json:"pid,omitempty"`
	ProcessTreePIDs       []int    `json:"processTreePids,omitempty"`
	WindowID              string   `json:"windowId,omitempty"`
	TabID                 string   `json:"tabId,omitempty"`
	TerminalID            string   `json:"terminalId,omitempty"`
	Title                 string   `json:"title"`
	CWD                   string   `json:"cwd"`
	IdentityStrategy      string   `json:"identityStrategy"`
	CapabilitiesSnapshot  []string `json:"capabilitiesSnapshot"`
	CanFocus              bool     `json:"canFocus"`
	CanClose              bool     `json:"canClose"`
	RequiresPlatformGrant bool     `json:"requiresPlatformGrant"`
}

// Session 表示一次助手终端会话
type Session struct {
	ID           string             `json:"id"`
	HarnessID    string             `json:"harnessId"`
	ProjectID    string             `json:"projectId,omitempty"`
	AdapterID    string             `json:"adapterId"`
	Title        string             `json:"title"`
	Command      string             `json:"command"`
	WorkingDir   string             `json:"workingDir"`
	WorktreePath string             `json:"worktreePath,omitempty"`
	Status       string             `json:"status"`
	CreatedAt    string             `json:"createdAt"`
	UpdatedAt    string             `json:"updatedAt"`
	PID          int                `json:"pid,omitempty"`
	SessionEnv   string             `json:"sessionEnv"`
	LaunchMode   string             `json:"launchMode"`
	FocusMode    string             `json:"focusMode"`
	CloseMode    string             `json:"closeMode"`
	Error        string             `json:"error,omitempty"`
	Ref          TerminalSessionRef `json:"ref"`
}

// Worktree 表示一个 Git worktree
type Worktree struct {
	ID         string `json:"id"`
	ProjectID  string `json:"projectId"`
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	BaseBranch string `json:"baseBranch,omitempty"`
	Head       string `json:"head,omitempty"`
	IsMain     bool   `json:"isMain"`
	Dirty      bool   `json:"dirty"`
	Exists     bool   `json:"exists"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// ActionResult 表示后端动作执行结果
type ActionResult struct {
	OK      bool   `json:"ok"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// WindowState 表示桌面窗口状态
type WindowState struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	AlwaysOnTop bool   `json:"alwaysOnTop"`
	Mode        string `json:"mode"`
}

// Bootstrap 表示前端启动时需要的一次性数据
type Bootstrap struct {
	Version          string            `json:"version"`
	Summary          AppSummary        `json:"summary"`
	ConfigPath       string            `json:"configPath"`
	Preferences      Preferences       `json:"preferences"`
	Harnesses        []Harness         `json:"harnesses"`
	Projects         []Project         `json:"projects"`
	TerminalProfiles []TerminalProfile `json:"terminalProfiles"`
	Adapters         []TerminalAdapter `json:"adapters"`
	Sessions         []Session         `json:"sessions"`
}
