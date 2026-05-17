package project

import (
	"siwap/internal/config"
	"siwap/internal/domain"
)

type Service struct {
	store *config.Store
}

func NewService(store *config.Store) *Service {
	return &Service{store: store}
}

func (s *Service) List() []domain.Project {
	return s.store.ListProjects()
}

func (s *Service) Get(id string) (domain.Project, bool) {
	return s.store.GetProject(id)
}

func (s *Service) Selected() (domain.Project, bool) {
	return s.store.SelectedProject()
}

func (s *Service) Add(path string, label string) (domain.Project, error) {
	return s.store.AddProject(path, label)
}

func (s *Service) Remove(id string) error {
	return s.store.RemoveProject(id)
}

func (s *Service) Select(id string) (domain.Project, error) {
	return s.store.SelectProject(id)
}

func (s *Service) SetDefault(id string) (domain.Project, error) {
	return s.store.SetDefaultProject(id)
}

func (s *Service) Reorder(ids []string) ([]domain.Project, error) {
	return s.store.ReorderProjects(ids)
}
