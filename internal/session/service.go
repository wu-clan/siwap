package session

import (
	"fmt"
	"sync"
	"time"

	"siwap/internal/domain"
	"siwap/internal/terminal"
)

// LaunchRequest 表示创建终端会话时需要的参数
type LaunchRequest struct {
	HarnessID     string            `json:"harnessId"`
	ProjectID     string            `json:"projectId"`
	AdapterID     string            `json:"adapterId"`
	Command       string            `json:"command"`
	WorkingDir    string            `json:"workingDir"`
	Title         string            `json:"title"`
	FlagOverrides map[string]string `json:"flagOverrides"`
	WorktreePath  string            `json:"worktreePath"`
}

// Service 管理应用内的会话状态
type Service struct {
	mu       sync.Mutex
	counter  int
	sessions []domain.Session
}

// NewService 创建会话服务
func NewService() *Service {
	return &Service{sessions: []domain.Session{}}
}

// List 返回当前所有会话的副本
func (s *Service) List() []domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneSessions(s.sessions)
}

// Get 根据会话 ID 查找会话
func (s *Service) Get(id string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, session := range s.sessions {
		if session.ID == id {
			return session, true
		}
	}
	return domain.Session{}, false
}

// Create 根据启动结果创建新的会话记录
func (s *Service) Create(req LaunchRequest, result terminal.LaunchResult, sessionEnv string) domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	now := time.Now().Format(time.RFC3339)
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("%s session %d", req.HarnessID, s.counter)
	}
	adapterID := result.Ref.AdapterID
	if adapterID == "" {
		adapterID = req.AdapterID
	}
	created := domain.Session{
		ID:           fmt.Sprintf("session-%d", s.counter),
		HarnessID:    req.HarnessID,
		ProjectID:    req.ProjectID,
		AdapterID:    adapterID,
		Title:        title,
		Command:      req.Command,
		WorkingDir:   req.WorkingDir,
		WorktreePath: req.WorktreePath,
		Status:       result.Status,
		CreatedAt:    now,
		UpdatedAt:    now,
		PID:          result.PID,
		SessionEnv:   sessionEnv,
		LaunchMode:   "foreground",
		FocusMode:    result.FocusMode,
		CloseMode:    result.CloseMode,
		Ref:          result.Ref,
	}
	s.sessions = append([]domain.Session{created}, s.sessions...)
	return created
}

// MarkError 创建失败状态的会话记录，便于前端展示错误
func (s *Service) MarkError(req LaunchRequest, err error, sessionEnv string) domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	now := time.Now().Format(time.RFC3339)
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("%s session %d", req.HarnessID, s.counter)
	}
	created := domain.Session{
		ID:           fmt.Sprintf("session-%d", s.counter),
		HarnessID:    req.HarnessID,
		ProjectID:    req.ProjectID,
		AdapterID:    req.AdapterID,
		Title:        title,
		Command:      req.Command,
		WorkingDir:   req.WorkingDir,
		WorktreePath: req.WorktreePath,
		Status:       "failed",
		CreatedAt:    now,
		UpdatedAt:    now,
		SessionEnv:   sessionEnv,
		Error:        err.Error(),
		LaunchMode:   "foreground",
	}
	s.sessions = append([]domain.Session{created}, s.sessions...)
	return created
}

// UpdateStatus 更新指定会话的状态和错误信息
func (s *Service) UpdateStatus(id string, status string, message string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.sessions {
		if s.sessions[i].ID == id {
			s.sessions[i].Status = status
			s.sessions[i].Error = message
			s.sessions[i].UpdatedAt = time.Now().Format(time.RFC3339)
			return s.sessions[i], true
		}
	}
	return domain.Session{}, false
}

// UpdateLaunch 使用重新启动后的结果更新会话
func (s *Service) UpdateLaunch(id string, result terminal.LaunchResult, sessionEnv string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.sessions {
		if s.sessions[i].ID == id {
			now := time.Now().Format(time.RFC3339)
			adapterID := result.Ref.AdapterID
			if adapterID == "" {
				adapterID = s.sessions[i].AdapterID
			}
			s.sessions[i].AdapterID = adapterID
			s.sessions[i].Status = result.Status
			s.sessions[i].UpdatedAt = now
			s.sessions[i].PID = result.PID
			s.sessions[i].SessionEnv = sessionEnv
			s.sessions[i].FocusMode = result.FocusMode
			s.sessions[i].CloseMode = result.CloseMode
			s.sessions[i].Ref = result.Ref
			s.sessions[i].Error = ""
			return s.sessions[i], true
		}
	}
	return domain.Session{}, false
}

// Remove 从会话列表中移除指定会话
func (s *Service) Remove(id string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, session := range s.sessions {
		if session.ID == id {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return session, true
		}
	}
	return domain.Session{}, false
}

// Clear 清空所有会话记录
func (s *Service) Clear() []domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := cloneSessions(s.sessions)
	s.sessions = []domain.Session{}
	return out
}

// cloneSessions 复制会话切片，避免外部修改内部状态
func cloneSessions(in []domain.Session) []domain.Session {
	out := make([]domain.Session, len(in))
	copy(out, in)
	return out
}
