package session

import (
	"fmt"
	"sync"
	"time"

	"siwap/internal/domain"
	"siwap/internal/terminal"
)

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

type Service struct {
	mu       sync.Mutex
	counter  int
	sessions []domain.Session
}

func NewService() *Service {
	return &Service{sessions: []domain.Session{}}
}

func (s *Service) List() []domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneSessions(s.sessions)
}

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

func (s *Service) Clear() []domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := cloneSessions(s.sessions)
	s.sessions = []domain.Session{}
	return out
}

func cloneSessions(in []domain.Session) []domain.Session {
	out := make([]domain.Session, len(in))
	copy(out, in)
	return out
}
