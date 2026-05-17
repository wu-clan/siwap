package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"siwap/internal/domain"
)

type LaunchRequest struct {
	AdapterID   string            `json:"adapterId"`
	Title       string            `json:"title"`
	Command     string            `json:"command"`
	WorkingDir  string            `json:"workingDir"`
	Environment map[string]string `json:"environment"`
	Background  bool              `json:"background"`
}

type LaunchResult struct {
	PID       int                       `json:"pid"`
	Status    string                    `json:"status"`
	Message   string                    `json:"message"`
	FocusMode string                    `json:"focusMode"`
	CloseMode string                    `json:"closeMode"`
	Ref       domain.TerminalSessionRef `json:"ref"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) List() []domain.TerminalAdapter {
	adapters := []domain.TerminalAdapter{
		autoAdapter(),
		ghosttyAdapter(),
		terminalAppAdapter(),
		windowsTerminalAdapter(),
		linuxDesktopAdapter(),
	}
	return adapters
}

func (s *Service) ListWithProfiles(profiles []domain.TerminalProfile) []domain.TerminalAdapter {
	adapters := s.List()
	for _, profile := range profiles {
		profile.Enabled = true
		installed := terminalProfileExists(profile.ExecutablePath)
		adapters = append(adapters, domain.TerminalAdapter{
			ID:         profile.ID,
			Label:      profile.Label,
			Platform:   firstNonEmpty(profile.Platform, runtime.GOOS),
			Executable: profile.ExecutablePath,
			Installed:  installed,
			Enabled:    profile.Enabled,
			Stability:  "custom profile",
			Confidence: boolConfidence(installed),
			Message:    availabilityMessage(installed, "Custom terminal profile is available.", "Custom terminal executable is missing."),
			Capabilities: []domain.TerminalCapability{
				capability("working-directory", "Working directory", true, "Uses the configured working directory flag or process cwd."),
				capability("environment", "Environment injection", true, "Inherits injected environment variables."),
				capability("title", "Session title", strings.Contains(profile.ArgumentTemplate, "{{title}}"), "Title support depends on the argument template."),
				capability("focus", "Focus session", false, "Custom profiles run in tracked-only focus mode."),
				capability("close", "Close session", true, "Best effort only: can interrupt the tracked launcher PID, not every terminal window."),
			},
		})
	}
	return adapters
}

func (s *Service) LaunchWithProfiles(req LaunchRequest, profiles []domain.TerminalProfile) (LaunchResult, error) {
	if strings.TrimSpace(req.Command) == "" {
		return LaunchResult{}, errors.New("command is required")
	}
	if req.WorkingDir == "" {
		wd, _ := os.UserHomeDir()
		req.WorkingDir = wd
	}
	if err := ensureDir(req.WorkingDir); err != nil {
		return LaunchResult{}, err
	}
	if req.Title == "" {
		req.Title = "Siwap Session"
	}
	adapterID := req.AdapterID
	if adapterID == "" || adapterID == "auto" {
		adapterID = s.bestAdapterID()
	}
	if adapterID == "" {
		return LaunchResult{}, errors.New("no installed terminal adapter available; configure a custom terminal profile")
	}
	for _, profile := range profiles {
		if profile.ID == adapterID {
			return s.launchProfile(req, profile)
		}
	}

	switch adapterID {
	case "ghostty":
		return s.launchGhostty(req)
	case "terminal-app":
		return s.launchMacTerminal(req)
	case "windows-terminal":
		return s.launchWindowsTerminal(req)
	case "linux-desktop":
		return s.launchLinuxTerminal(req)
	default:
		return LaunchResult{}, fmt.Errorf("unknown terminal adapter: %s", adapterID)
	}
}

func (s *Service) launchProfile(req LaunchRequest, profile domain.TerminalProfile) (LaunchResult, error) {
	if profile.ExecutablePath == "" {
		return LaunchResult{}, fmt.Errorf("terminal profile %s has no executable", profile.Label)
	}
	if !terminalProfileExists(profile.ExecutablePath) {
		return LaunchResult{}, fmt.Errorf("terminal profile executable does not exist: %s", profile.ExecutablePath)
	}
	args := renderArgs(profile.ArgumentTemplate, req)
	args = append(renderWorkingDirArgs(profile.WorkingDirFlag, req), args...)
	cmd := terminalProfileCommand(profile.ExecutablePath, args)
	cmd.Env = mergeEnv(req.Environment)
	cmd.Dir = req.WorkingDir
	if err := cmd.Start(); err != nil {
		return LaunchResult{}, err
	}
	result := launchResult(req, profile.ID, cmd.Process.Pid, "launched", "Launched with custom terminal profile "+profile.Label+".", false, true)
	result.Ref.IdentityStrategy = "custom profile tracked launcher pid + session title"
	return result, nil
}

func (s *Service) Focus(session domain.Session) domain.ActionResult {
	switch session.AdapterID {
	case "ghostty":
		if runtime.GOOS == "darwin" {
			if session.Ref.WindowID != "" && session.Ref.TabID != "" && session.Ref.TerminalID != "" {
				return focusDarwinGhosttyTerminal(session)
			}
			if !darwinProcessExists("Ghostty") {
				return domain.ActionResult{OK: false, Status: "gone", Message: "Ghostty is not running; reopening session terminal."}
			}
			return focusDarwinApp("Ghostty", "Ghostty")
		}
	case "terminal-app":
		if runtime.GOOS == "darwin" {
			if !darwinProcessExists("Terminal") {
				return domain.ActionResult{OK: false, Status: "gone", Message: "Terminal is not running; reopening session terminal."}
			}
			if session.Ref.WindowID != "" {
				return focusDarwinTerminalAppWindow(session.Ref.WindowID)
			}
			return focusDarwinApp("Terminal", "Terminal")
		}
	case "windows-terminal":
		if runtime.GOOS == "windows" {
			_ = exec.Command("powershell", "-NoProfile", "-Command", `(New-Object -ComObject WScript.Shell).AppActivate('Windows Terminal')`).Run()
			return domain.ActionResult{OK: true, Status: "best-effort", Message: "Requested Windows Terminal focus."}
		}
	}
	if session.PID > 0 {
		if IsProcessAlive(session.PID) {
			return domain.ActionResult{OK: true, Status: "tracked", Message: fmt.Sprintf("Session launcher PID %d is tracked; native focus is unavailable for this adapter.", session.PID)}
		}
		return domain.ActionResult{OK: true, Status: "unverified", Message: fmt.Sprintf("Tracked launcher PID %d has exited; the terminal window may still be open, so Siwap will not reopen it automatically.", session.PID)}
	}
	return domain.ActionResult{OK: false, Status: "unsupported", Message: "Native focus is not available for this adapter on this platform."}
}

func (s *Service) Close(session domain.Session) domain.ActionResult {
	if runtime.GOOS == "darwin" {
		if result := closeDarwinTerminalWindow(session); result.Status != "unsupported" {
			return result
		}
	}
	if session.PID <= 0 {
		return domain.ActionResult{OK: true, Status: "removed", Message: "No process was tracked; session removed from Siwap only."}
	}
	if !IsProcessAlive(session.PID) {
		return domain.ActionResult{OK: true, Status: "gone", Message: fmt.Sprintf("Tracked launcher PID %d is already gone; session removed from Siwap only.", session.PID)}
	}
	if runtime.GOOS == "darwin" {
		exitFullScreenBestEffort(session.AdapterID)
	}
	process, err := os.FindProcess(session.PID)
	if err != nil {
		return domain.ActionResult{OK: true, Status: "gone", Message: "Process is already gone."}
	}
	if err := process.Signal(os.Interrupt); err != nil {
		_ = process.Kill()
		return domain.ActionResult{OK: true, Status: "killed", Message: fmt.Sprintf("Sent kill signal to PID %d after interrupt failed.", session.PID)}
	}
	return domain.ActionResult{OK: true, Status: "closing", Message: fmt.Sprintf("Sent interrupt to PID %d.", session.PID)}
}

func (s *Service) bestAdapterID() string {
	for _, adapter := range s.List() {
		if adapter.ID == "auto" {
			continue
		}
		if adapter.Installed {
			return adapter.ID
		}
	}
	return ""
}

func (s *Service) launchGhostty(req LaunchRequest) (LaunchResult, error) {
	if runtime.GOOS == "darwin" && appExists("/Applications/Ghostty.app") {
		return s.launchDarwinGhostty(req)
	}
	if !CommandExists("ghostty") {
		return LaunchResult{}, errors.New("Ghostty CLI is not installed")
	}
	env := mergeEnv(req.Environment)
	shell := userShell()
	cmd := exec.Command("ghostty", ghosttyLaunchArgs(req, shell)...)
	cmd.Env = env
	cmd.Dir = req.WorkingDir
	if err := cmd.Start(); err != nil {
		return LaunchResult{}, err
	}
	return launchResult(req, "ghostty", cmd.Process.Pid, "launched", "Launched with Ghostty CLI.", true, true), nil
}

type ghosttySurface struct {
	WindowID         string `json:"windowId"`
	TabID            string `json:"tabId"`
	TerminalID       string `json:"terminalId"`
	WorkingDirectory string `json:"workingDirectory"`
}

type terminalAppSurface struct {
	WindowID string `json:"windowId"`
	TabID    string `json:"tabId"`
	Title    string `json:"title"`
}

func (s *Service) launchDarwinGhostty(req LaunchRequest) (LaunchResult, error) {
	initialInput := terminalShellCommand(req)
	script := fmt.Sprintf(`
%s

tell application "Ghostty"
	set existingWindowIds to {}
	try
		set winCountBefore to count of windows
		repeat with winIdx from 1 to winCountBefore
			try
				copy (id of (window winIdx) as text) to end of existingWindowIds
			end try
		end repeat
	end try

	set cfg to new surface configuration
	set initial working directory of cfg to %s
	%s
	set initial input of cfg to %s & linefeed
	set createdWin to new window with configuration cfg

	set targetWin to missing value
	set targetTab to missing value
	set targetTerm to missing value

	repeat 30 times
		try
			set candidateId to id of createdWin as text
			if existingWindowIds does not contain candidateId then
				set targetWin to createdWin
				set targetTab to selected tab of createdWin
				set targetTerm to focused terminal of targetTab
				exit repeat
			end if
		end try
		delay 0.05
	end repeat

	if targetTerm is missing value then
		try
			set winCountAfter to count of windows
			repeat with winIdx from 1 to winCountAfter
				try
					set winRef to window winIdx
					set winId to id of winRef as text
					if existingWindowIds does not contain winId then
						set targetWin to winRef
						set targetTab to selected tab of winRef
						set targetTerm to focused terminal of targetTab
						exit repeat
					end if
				end try
			end repeat
		end try
	end if

	if targetTerm is missing value then error "Siwap could not verify the newly created Ghostty terminal."

	select tab targetTab
	focus targetTerm

	set output to "{"
	set output to output & "\"windowId\":" & my jsonString(id of targetWin)
	set output to output & ",\"tabId\":" & my jsonString(id of targetTab)
	set output to output & ",\"terminalId\":" & my jsonString(id of targetTerm)
	set output to output & ",\"workingDirectory\":" & my jsonString(working directory of targetTerm)
	set output to output & "}"
	return output
end tell`, appleScriptJSONHandlers, appleScriptLiteral(req.WorkingDir), ghosttyEnvironmentClause(req.Environment), appleScriptLiteral(initialInput))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return LaunchResult{}, err
	}
	var surface ghosttySurface
	if err := json.Unmarshal([]byte(out), &surface); err != nil {
		return LaunchResult{}, err
	}
	result := launchResult(req, "ghostty", darwinProcessPID("Ghostty"), "launched", "Launched with Ghostty AppleScript.", true, true)
	result.Ref.WindowID = surface.WindowID
	result.Ref.TabID = surface.TabID
	result.Ref.TerminalID = surface.TerminalID
	if surface.WorkingDirectory != "" {
		result.Ref.CWD = surface.WorkingDirectory
	}
	result.Ref.IdentityStrategy = "ghostty window id + tab id + terminal id"
	return result, nil
}

func ghosttyLaunchArgs(req LaunchRequest, shell string) []string {
	return []string{
		"--working-directory=" + req.WorkingDir,
		"--title=" + req.Title,
		"-e",
		shell,
		"-lc",
		req.Command,
	}
}

func (s *Service) launchMacTerminal(req LaunchRequest) (LaunchResult, error) {
	if runtime.GOOS != "darwin" {
		return LaunchResult{}, errors.New("macOS Terminal.app adapter is only available on macOS")
	}
	command := terminalShellCommand(req)
	script := fmt.Sprintf(`
%s

tell application "Terminal"
	set existingWindowIds to {}
	try
		set winCountBefore to count of windows
		repeat with winIdx from 1 to winCountBefore
			try
				copy (id of (window winIdx) as text) to end of existingWindowIds
			end try
		end repeat
	end try

	set createdTab to do script %s
	activate

	set targetWin to missing value
	repeat 30 times
		try
			set winCountAfter to count of windows
			repeat with winIdx from 1 to winCountAfter
				try
					set winRef to window winIdx
					set winId to id of winRef as text
					if existingWindowIds does not contain winId then
						set targetWin to winRef
						exit repeat
					end if
				end try
			end repeat
		end try
		if targetWin is not missing value then exit repeat
		delay 0.05
	end repeat

	if targetWin is missing value then
		try
			set targetWin to front window
		end try
	end if
	if targetWin is missing value then error "Siwap could not verify the newly created Terminal window."

	set targetTabId to ""
	try
		set targetTabId to id of createdTab as text
	end try

	set output to "{"
	set output to output & "\"windowId\":" & my jsonString(id of targetWin)
	set output to output & ",\"tabId\":" & my jsonString(targetTabId)
	set output to output & ",\"title\":" & my jsonString(name of targetWin)
	set output to output & "}"
	return output
end tell`, appleScriptJSONHandlers, strconv.Quote(command))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return LaunchResult{}, err
	}
	var surface terminalAppSurface
	if err := json.Unmarshal([]byte(out), &surface); err != nil {
		return LaunchResult{}, err
	}
	result := launchResult(req, "terminal-app", darwinProcessPID("Terminal"), "launched", "Launched with macOS Terminal AppleScript.", true, true)
	result.Ref.WindowID = surface.WindowID
	result.Ref.TabID = surface.TabID
	result.Ref.IdentityStrategy = "terminal.app window id + session title"
	return result, nil
}

func (s *Service) launchWindowsTerminal(req LaunchRequest) (LaunchResult, error) {
	if runtime.GOOS != "windows" {
		return LaunchResult{}, errors.New("Windows Terminal adapter is only available on Windows")
	}
	if !CommandExists("wt") {
		return LaunchResult{}, errors.New("Windows Terminal executable wt was not found")
	}
	cmd := exec.Command("wt", windowsTerminalArgs(req, userShell())...)
	cmd.Env = mergeEnv(req.Environment)
	cmd.Dir = req.WorkingDir
	if err := cmd.Start(); err != nil {
		return LaunchResult{}, err
	}
	return launchResult(req, "windows-terminal", cmd.Process.Pid, "launched", "Launched with Windows Terminal.", true, false), nil
}

func windowsTerminalArgs(req LaunchRequest, shell string) []string {
	return []string{"-w", "0", "new-tab", "--title", req.Title, "--startingDirectory", req.WorkingDir, shell, "/K", req.Command}
}

func (s *Service) launchLinuxTerminal(req LaunchRequest) (LaunchResult, error) {
	if runtime.GOOS != "linux" {
		return LaunchResult{}, errors.New("Linux desktop terminal adapter is only available on Linux")
	}
	candidates := linuxTerminalCandidates(req, userShell())
	for _, candidate := range candidates {
		if !CommandExists(candidate[0]) {
			continue
		}
		cmd := exec.Command(candidate[0], candidate[1:]...)
		cmd.Env = mergeEnv(req.Environment)
		cmd.Dir = req.WorkingDir
		if err := cmd.Start(); err == nil {
			return launchResult(req, "linux-desktop", cmd.Process.Pid, "launched", "Launched with Linux desktop terminal.", true, false), nil
		}
	}
	return LaunchResult{}, errors.New("no supported Linux desktop terminal command was found")
}

func linuxTerminalCandidates(req LaunchRequest, shell string) [][]string {
	return [][]string{
		{"x-terminal-emulator", "-e", shell, "-lc", req.Command},
		{"gnome-terminal", "--working-directory", req.WorkingDir, "--title", req.Title, "--", shell, "-lc", req.Command},
		{"konsole", "--workdir", req.WorkingDir, "-p", "tabtitle=" + req.Title, "-e", shell, "-lc", req.Command},
		{"xterm", "-T", req.Title, "-e", shell, "-lc", req.Command},
	}
}

func launchResult(req LaunchRequest, adapterID string, pid int, status string, message string, canFocus bool, canClose bool) LaunchResult {
	return LaunchResult{
		PID:       pid,
		Status:    status,
		Message:   message,
		FocusMode: focusMode(canFocus),
		CloseMode: closeMode(canClose, pid),
		Ref: domain.TerminalSessionRef{
			AdapterID:            adapterID,
			Platform:             runtime.GOOS,
			PID:                  pid,
			Title:                req.Title,
			CWD:                  req.WorkingDir,
			IdentityStrategy:     "session env > pid > cwd + createdAt",
			CapabilitiesSnapshot: []string{"working-directory", "environment", "title", "process"},
			CanFocus:             canFocus,
			CanClose:             canClose,
		},
	}
}

func autoAdapter() domain.TerminalAdapter {
	return domain.TerminalAdapter{
		ID:         "auto",
		Label:      "Auto select",
		Platform:   runtime.GOOS,
		Installed:  true,
		Enabled:    true,
		Stability:  "adaptive",
		Confidence: "high",
		Message:    "Chooses the best installed adapter for this platform.",
		Capabilities: []domain.TerminalCapability{
			capability("working-directory", "Working directory", true, "Launch in a selected project or worktree."),
			capability("environment", "Environment injection", true, "Inject SIWAP_SESSION_ID and metadata."),
			capability("focus", "Focus session", runtime.GOOS != "linux", "Best effort focus through platform APIs."),
		},
	}
}

func ghosttyAdapter() domain.TerminalAdapter {
	installed := CommandExists("ghostty") || appExists("/Applications/Ghostty.app")
	return domain.TerminalAdapter{
		ID:         "ghostty",
		Label:      "Ghostty",
		Platform:   runtime.GOOS,
		Executable: executablePath("ghostty", "/Applications/Ghostty.app"),
		Installed:  installed,
		Enabled:    true,
		Stability:  "high on macOS, fallback elsewhere",
		Confidence: boolConfidence(installed),
		Message:    availabilityMessage(installed, "Ghostty CLI or app detected.", "Ghostty is not installed; auto mode will fall back."),
		Capabilities: []domain.TerminalCapability{
			capability("working-directory", "Working directory", true, "Launch a terminal session in a project directory."),
			capability("environment", "Environment injection", true, "Inject SIWAP_SESSION_ID for process and port attribution."),
			capability("title", "Session title", true, "Assign a traceable title."),
			capability("focus", "Focus session", runtime.GOOS == "darwin", "macOS can activate Ghostty; other platforms fall back."),
			capability("close", "Close session", true, "Close tracked root PID when available."),
			capability("window-arrange", "Window arrangement", runtime.GOOS == "darwin", "Requires macOS Accessibility permissions for reliable bounds changes."),
			capability("space-switch", "Space switching", false, "Private macOS APIs are intentionally not used in the MVP."),
		},
	}
}

func terminalAppAdapter() domain.TerminalAdapter {
	installed := runtime.GOOS == "darwin" && appExists("/System/Applications/Utilities/Terminal.app")
	return domain.TerminalAdapter{
		ID:         "terminal-app",
		Label:      "macOS Terminal.app",
		Platform:   runtime.GOOS,
		Executable: "/System/Applications/Utilities/Terminal.app",
		Installed:  installed,
		Enabled:    true,
		Stability:  "macOS fallback",
		Confidence: boolConfidence(installed),
		Message:    availabilityMessage(installed, "Terminal.app AppleScript launch is available.", "Terminal.app adapter is macOS-only."),
		Capabilities: []domain.TerminalCapability{
			capability("working-directory", "Working directory", installed, "Uses shell cd before running the command."),
			capability("environment", "Environment injection", installed, "Exports session variables in shell script."),
			capability("focus", "Focus session", installed, "Activates Terminal.app."),
			capability("close", "Close session", installed, "Closes the tracked Terminal.app window when launched by Siwap."),
		},
	}
}

func windowsTerminalAdapter() domain.TerminalAdapter {
	installed := runtime.GOOS == "windows" && CommandExists("wt")
	return domain.TerminalAdapter{
		ID:         "windows-terminal",
		Label:      "Windows Terminal",
		Platform:   runtime.GOOS,
		Executable: executablePath("wt", ""),
		Installed:  installed,
		Enabled:    true,
		Stability:  "planned native, launch implemented",
		Confidence: boolConfidence(installed),
		Message:    availabilityMessage(installed, "wt.exe detected.", "Windows Terminal is only available on Windows."),
		Capabilities: []domain.TerminalCapability{
			capability("working-directory", "Working directory", installed, "Uses wt.exe -d."),
			capability("environment", "Environment injection", installed, "Inherits injected process environment."),
			capability("focus", "Focus session", installed, "Best effort AppActivate."),
			capability("close", "Close session", false, "PID tracking is best effort for wt.exe."),
		},
	}
}

func linuxDesktopAdapter() domain.TerminalAdapter {
	installed := runtime.GOOS == "linux" && (CommandExists("x-terminal-emulator") || CommandExists("gnome-terminal") || CommandExists("konsole") || CommandExists("xterm"))
	return domain.TerminalAdapter{
		ID:         "linux-desktop",
		Label:      "Linux desktop terminal",
		Platform:   runtime.GOOS,
		Executable: executablePath("x-terminal-emulator", executablePath("gnome-terminal", executablePath("konsole", executablePath("xterm", "")))),
		Installed:  installed,
		Enabled:    true,
		Stability:  "desktop dependent",
		Confidence: boolConfidence(installed),
		Message:    availabilityMessage(installed, "A Linux desktop terminal command was detected.", "No Linux desktop terminal command detected."),
		Capabilities: []domain.TerminalCapability{
			capability("working-directory", "Working directory", installed, "Depends on terminal implementation."),
			capability("environment", "Environment injection", installed, "Inherits injected process environment."),
			capability("focus", "Focus session", false, "Wayland/X11 focus requires optional tools."),
			capability("close", "Close session", true, "Can close tracked launcher PID when available."),
		},
	}
}

func terminalShellCommand(req LaunchRequest) string {
	parts := []string{"printf '\\033]0;%s\\007' " + shellQuote(req.Title), "cd " + shellQuote(req.WorkingDir)}
	for key, value := range req.Environment {
		parts = append(parts, "export "+key+"="+shellQuote(value))
	}
	parts = append(parts, "clear")
	parts = append(parts, req.Command)
	return strings.Join(parts, "; ")
}

func mergeEnv(values map[string]string) []string {
	env := os.Environ()
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	return env
}

func focusMode(canFocus bool) string {
	if canFocus {
		return "native-or-platform"
	}
	return "tracked-only"
}

func closeMode(canClose bool, pid int) string {
	if canClose && pid > 0 {
		return "pid-interrupt"
	}
	return "remove-only"
}

func userShell() string {
	if runtime.GOOS == "windows" {
		if ComSpec := os.Getenv("ComSpec"); ComSpec != "" {
			return ComSpec
		}
		return "cmd"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}
	return nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func capability(key, label string, supported bool, description string) domain.TerminalCapability {
	return domain.TerminalCapability{Key: key, Label: label, Supported: supported, Description: description}
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func appExists(path string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func executablePath(name string, fallback string) string {
	if name != "" {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return fallback
}

func boolConfidence(ok bool) string {
	if ok {
		return "high"
	}
	return "unavailable"
}

func availabilityMessage(ok bool, yes string, no string) string {
	if ok {
		return yes
	}
	return no
}

func runAppleScript(script string) domain.ActionResult {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return domain.ActionResult{OK: false, Status: "failed", Message: strings.TrimSpace(string(out)) + " " + err.Error()}
	}
	return domain.ActionResult{OK: true, Status: "ok", Message: "AppleScript completed."}
}

func runAppleScriptOutput(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func appleScriptLiteral(value string) string {
	escaped := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
	).Replace(value)
	return `"` + escaped + `"`
}

func ghosttyEnvironmentClause(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	items := make([]string, 0, len(values))
	for key, value := range values {
		items = append(items, appleScriptLiteral(key+"="+value))
	}
	return "set environment variables of cfg to {" + strings.Join(items, ", ") + "}"
}

const appleScriptJSONHandlers = `
on replaceText(findText, replacementText, sourceText)
	set AppleScript's text item delimiters to findText
	set parts to text items of sourceText
	set AppleScript's text item delimiters to replacementText
	set updatedText to parts as text
	set AppleScript's text item delimiters to ""
	return updatedText
end replaceText

on jsonString(valueText)
	if valueText is missing value then set valueText to ""
	set escapedText to valueText as text
	set escapedText to my replaceText("\\", "\\\\", escapedText)
	set escapedText to my replaceText(quote, "\\" & quote, escapedText)
	set escapedText to my replaceText(linefeed, "\n", escapedText)
	set escapedText to my replaceText(return, "\r", escapedText)
	return quote & escapedText & quote
end jsonString
`

func focusDarwinApp(appName string, processName string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application %s to activate
tell application "System Events"
	if exists process %s then
		set visible of process %s to true
		tell process %s
			if (count of windows) > 0 then
				try
					set value of attribute "AXMinimized" of front window to false
				end try
				try
					perform action "AXRaise" of front window
				end try
			end if
		end tell
	end if
end tell`, strconv.Quote(appName), strconv.Quote(processName), strconv.Quote(processName), strconv.Quote(processName))
	result := runAppleScript(script)
	if result.OK {
		result.Status = "focused"
		result.Message = "Requested focus, unhide, unminimize, and raise for " + appName + "."
	}
	return result
}

func darwinProcessExists(processName string) bool {
	return exec.Command("pgrep", "-x", processName).Run() == nil
}

func darwinProcessPID(processName string) int {
	out, err := exec.Command("pgrep", "-nx", processName).Output()
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return pid
}

func focusDarwinGhosttyTerminal(session domain.Session) domain.ActionResult {
	script := fmt.Sprintf(`
%s

tell application "Ghostty"
	set expectedWindowId to %s
	set expectedTabId to %s
	set expectedTerminalId to %s

	set targetWin to missing value
	set winCount to 0
	try
		set winCount to count of windows
	end try
	repeat with winIdx from 1 to winCount
		try
			set winRef to window winIdx
			if (id of winRef as text) is expectedWindowId then
				set targetWin to winRef
				exit repeat
			end if
		end try
	end repeat
	if targetWin is missing value then error "Siwap could not find the managed Ghostty window."

	set targetTab to missing value
	set tabCount to 0
	try
		set tabCount to count of tabs of targetWin
	end try
	repeat with tabIdx from 1 to tabCount
		try
			set tabRef to tab tabIdx of targetWin
			if (id of tabRef as text) is expectedTabId then
				set targetTab to tabRef
				exit repeat
			end if
		end try
	end repeat
	if targetTab is missing value then error "Siwap could not find the managed Ghostty tab."

	set targetTerm to missing value
	set termCount to 0
	try
		set termCount to count of terminals of targetTab
	end try
	repeat with termIdx from 1 to termCount
		try
			set termRef to terminal termIdx of targetTab
			if (id of termRef as text) is expectedTerminalId then
				set targetTerm to termRef
				exit repeat
			end if
		end try
	end repeat
	if targetTerm is missing value then error "Siwap could not find the managed Ghostty terminal."

	try
		set visible of targetWin to true
	end try
	try
		set miniaturized of targetWin to false
	end try

	select tab targetTab
	focus targetTerm

	set output to "{"
	set output to output & "\"windowId\":" & my jsonString(id of targetWin)
	set output to output & ",\"tabId\":" & my jsonString(id of targetTab)
	set output to output & ",\"terminalId\":" & my jsonString(id of targetTerm)
	set output to output & ",\"workingDirectory\":" & my jsonString(working directory of targetTerm)
	set output to output & "}"
	return output
end tell`, appleScriptJSONHandlers, appleScriptLiteral(session.Ref.WindowID), appleScriptLiteral(session.Ref.TabID), appleScriptLiteral(session.Ref.TerminalID))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: err.Error()}
	}
	var surface ghosttySurface
	if err := json.Unmarshal([]byte(out), &surface); err != nil {
		return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
	}
	return domain.ActionResult{OK: true, Status: "focused", Message: "Focused Ghostty session."}
}

func focusDarwinTerminalAppWindow(windowID string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application "Terminal"
	set expectedWindowId to %s
	set targetWin to missing value
	try
		set winCount to count of windows
		repeat with winIdx from 1 to winCount
			try
				set winRef to window winIdx
				if (id of winRef as text) is expectedWindowId then
					set targetWin to winRef
					exit repeat
				end if
			end try
		end repeat
	end try
	if targetWin is missing value then error "Siwap could not find the managed Terminal window."
	set index of targetWin to 1
	activate
end tell
tell application "System Events"
	if exists process "Terminal" then
		tell process "Terminal"
			if (count of windows) > 0 then
				try
					perform action "AXRaise" of front window
				end try
			end if
		end tell
	end if
end tell
return "focused"`, appleScriptLiteral(windowID))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: err.Error()}
	}
	if strings.TrimSpace(out) == "focused" {
		return domain.ActionResult{OK: true, Status: "focused", Message: "Focused Terminal session."}
	}
	return domain.ActionResult{OK: false, Status: "failed", Message: out}
}

func closeDarwinTerminalWindow(session domain.Session) domain.ActionResult {
	if session.AdapterID == "ghostty" {
		if session.Ref.WindowID != "" && session.Ref.TabID != "" && session.Ref.TerminalID != "" {
			return closeDarwinGhosttyTerminal(session.Ref.WindowID, session.Ref.TabID, session.Ref.TerminalID)
		}
		if session.Ref.WindowID != "" {
			return closeDarwinGhosttyWindow(session.Ref.WindowID)
		}
	}
	if session.AdapterID == "terminal-app" && session.Ref.WindowID != "" {
		return closeDarwinTerminalAppWindow(session.Ref.WindowID, "")
	}
	title := strings.TrimSpace(session.Ref.Title)
	if title == "" {
		title = strings.TrimSpace(session.SessionEnv)
	}
	if title == "" {
		return domain.ActionResult{OK: false, Status: "unsupported", Message: "No terminal window title is tracked for this session."}
	}
	switch session.AdapterID {
	case "terminal-app":
		return closeDarwinTerminalAppWindow("", title)
	case "ghostty":
		return closeDarwinGUIWindow("Ghostty", title)
	default:
		return domain.ActionResult{OK: false, Status: "unsupported", Message: "Native window close is not available for this adapter."}
	}
}

func closeDarwinGhosttyTerminal(windowID string, tabID string, terminalID string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application "Ghostty"
	set expectedWindowId to %s
	set expectedTabId to %s
	set expectedTerminalId to %s

	set targetTerm to missing value
	set targetTab to missing value
	try
		set targetTerm to terminal id expectedTerminalId
	end try

	if targetTerm is not missing value then
		focus targetTerm
		set targetTab to selected tab of front window
	else
		set targetWin to missing value
		set winCount to 0
		try
			set winCount to count of windows
		end try
		repeat with winIdx from 1 to winCount
			try
				set winRef to window winIdx
				if (id of winRef as text) is expectedWindowId then
					set targetWin to winRef
					exit repeat
				end if
			end try
		end repeat
		if targetWin is missing value then error "Siwap could not find the managed Ghostty window."

		set tabCount to 0
		try
			set tabCount to count of tabs of targetWin
		end try
		repeat with tabIdx from 1 to tabCount
			try
				set tabRef to tab tabIdx of targetWin
				if (id of tabRef as text) is expectedTabId then
					set targetTab to tabRef
					exit repeat
				end if
			end try
		end repeat
		if targetTab is missing value then error "Siwap could not find the managed Ghostty tab."

		set termCount to 0
		try
			set termCount to count of terminals of targetTab
		end try
		repeat with termIdx from 1 to termCount
			try
				set termRef to terminal termIdx of targetTab
				if (id of termRef as text) is expectedTerminalId then
					set targetTerm to termRef
					exit repeat
				end if
			end try
		end repeat
		if targetTerm is missing value then error "Siwap could not find the managed Ghostty terminal."
	end if

	if targetTab is missing value then error "Siwap could not find the managed Ghostty tab."

	close tab targetTab
	return "closed"
end tell`, appleScriptLiteral(windowID), appleScriptLiteral(tabID), appleScriptLiteral(terminalID))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: err.Error()}
	}
	if strings.TrimSpace(out) == "closed" {
		return domain.ActionResult{OK: true, Status: "closed", Message: "Closed Ghostty terminal for session."}
	}
	return domain.ActionResult{OK: false, Status: "failed", Message: out}
}

func closeDarwinTerminalAppWindow(windowID string, title string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application "Terminal"
	set expectedWindowId to %s
	if expectedWindowId is not "" then
		repeat with w in windows
			if (id of w as text) is expectedWindowId then
				set index of w to 1
				activate
				my clickTerminalFrontWindowCloseButton()
				delay 0.1
				my terminateTerminalCloseConfirmation()
				return "closed"
			end if
		end repeat
		return "not-found"
	end if
	repeat with w in windows
		if (name of w as text) contains %s then
			set index of w to 1
			activate
			my clickTerminalFrontWindowCloseButton()
			delay 0.1
			my terminateTerminalCloseConfirmation()
			return "closed"
		end if
	end repeat
end tell
return "not-found"

on clickTerminalFrontWindowCloseButton()
	tell application "System Events"
		if not (exists process "Terminal") then return
		tell process "Terminal"
			if (count of windows) is 0 then return
			try
				perform action "AXRaise" of front window
			end try
			try
				click button 1 of front window
			end try
		end tell
	end tell
end clickTerminalFrontWindowCloseButton

on terminateTerminalCloseConfirmation()
	tell application "System Events"
		if not (exists process "Terminal") then return
		tell process "Terminal"
			repeat 10 times
				try
					if (count of sheets of front window) > 0 then
						tell sheet 1 of front window
							if exists button "Terminate" then
								click button "Terminate"
								return
							end if
							if exists button "Close" then
								click button "Close"
								return
							end if
							if exists button "OK" then
								click button "OK"
								return
							end if
							key code 36
							return
						end tell
					end if
				end try
				delay 0.05
			end repeat
		end tell
	end tell
end terminateTerminalCloseConfirmation`, appleScriptLiteral(windowID), strconv.Quote(title))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
	}
	if strings.TrimSpace(out) == "closed" {
		return domain.ActionResult{OK: true, Status: "closed", Message: "Closed Terminal window for session."}
	}
	return domain.ActionResult{OK: false, Status: "unsupported", Message: "No Terminal window matched this session."}
}

func closeDarwinGhosttyWindow(windowID string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application "Ghostty"
	set expectedWindowId to %s
	set targetWin to missing value
	try
		set targetWin to window id expectedWindowId
	end try

	set winCount to 0
	try
		set winCount to count of windows
	end try
	if targetWin is missing value then
		repeat with winIdx from 1 to winCount
			try
				set winRef to window winIdx
				if (id of winRef as text) is expectedWindowId then
					set targetWin to winRef
					exit repeat
				end if
			end try
		end repeat
	end if

	if targetWin is missing value then error "Siwap could not find the managed Ghostty window."
	close window targetWin
	return "closed"
end tell`, appleScriptLiteral(windowID))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "missing", Message: err.Error()}
	}
	if strings.TrimSpace(out) == "closed" {
		return domain.ActionResult{OK: true, Status: "closed", Message: "Closed Ghostty window for session."}
	}
	return domain.ActionResult{OK: false, Status: "failed", Message: out}
}

func closeDarwinGUIWindow(processName string, title string) domain.ActionResult {
	script := fmt.Sprintf(`
tell application "System Events"
	if not (exists process %s) then return "not-running"
	tell process %s
		repeat with w in windows
			if (name of w as text) contains %s then
				perform action "AXRaise" of w
				click button 1 of w
				return "closed"
			end if
		end repeat
	end tell
end tell
return "not-found"`, strconv.Quote(processName), strconv.Quote(processName), strconv.Quote(title))
	out, err := runAppleScriptOutput(script)
	if err != nil {
		return domain.ActionResult{OK: false, Status: "failed", Message: err.Error()}
	}
	switch strings.TrimSpace(out) {
	case "closed":
		return domain.ActionResult{OK: true, Status: "closed", Message: "Closed " + processName + " window for session."}
	case "not-running":
		return domain.ActionResult{OK: true, Status: "gone", Message: processName + " is already closed."}
	default:
		return domain.ActionResult{OK: false, Status: "unsupported", Message: "No " + processName + " window matched this session."}
	}
}

func exitFullScreenBestEffort(adapterID string) {
	processName := ""
	switch adapterID {
	case "ghostty":
		processName = "Ghostty"
	case "terminal-app":
		processName = "Terminal"
	default:
		return
	}
	script := fmt.Sprintf(`
tell application "System Events"
	if exists process %s then
		tell process %s
			if (count of windows) > 0 then
				try
					set value of attribute "AXFullScreen" of front window to false
				end try
			end if
		end tell
	end if
end tell`, strconv.Quote(processName), strconv.Quote(processName))
	_, _ = runAppleScriptOutput(script)
}

func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func SessionID() string {
	return fmt.Sprintf("siwap-%d", time.Now().UnixNano())
}

func WorktreeSafeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, string(filepath.Separator), "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	if name == "" {
		return "session"
	}
	return name
}

func renderArgs(template string, req LaunchRequest) []string {
	if strings.TrimSpace(template) == "" {
		template = "{{command}}"
	}
	replacements := map[string]string{
		"{{command}}": req.Command,
		"{{cwd}}":     req.WorkingDir,
		"{{title}}":   req.Title,
	}
	tokens := splitArgs(template)
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		for key, value := range replacements {
			token = strings.ReplaceAll(token, key, value)
		}
		if token != "" {
			out = append(out, token)
		}
	}
	return out
}

func renderWorkingDirArgs(flag string, req LaunchRequest) []string {
	flag = strings.TrimSpace(flag)
	if flag == "" {
		return nil
	}
	if strings.Contains(flag, "{{cwd}}") {
		return renderArgs(flag, req)
	}
	if strings.HasSuffix(flag, "=") {
		return []string{flag + req.WorkingDir}
	}
	return []string{flag, req.WorkingDir}
}

func splitArgs(input string) []string {
	var out []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' || r == '\n' {
			if current.Len() > 0 {
				out = append(out, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		out = append(out, current.String())
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func terminalProfileExists(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if fileExists(path) {
		return true
	}
	return runtime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(path), ".app") && appExists(path)
}

func terminalProfileCommand(path string, args []string) *exec.Cmd {
	if runtime.GOOS == "darwin" && strings.HasSuffix(strings.ToLower(strings.TrimSpace(path)), ".app") {
		openArgs := append([]string{path, "--args"}, args...)
		return exec.Command("open", openArgs...)
	}
	return exec.Command(path, args...)
}

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
