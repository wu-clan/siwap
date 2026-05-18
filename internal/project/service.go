package project

import (
	"siwap/internal/config"
	"siwap/internal/domain"
)

// Service 提供项目配置的读取和变更能力
type Service struct {
	store *config.Store
}

// NewService 创建项目服务
func NewService(store *config.Store) *Service {
	return &Service{store: store}
}

// List 返回项目列表
func (s *Service) List() []domain.Project {
	return s.store.ListProjects()
}

// Get 根据项目 ID 查找项目
func (s *Service) Get(id string) (domain.Project, bool) {
	return s.store.GetProject(id)
}

// Selected 返回当前选中的项目
func (s *Service) Selected() (domain.Project, bool) {
	return s.store.SelectedProject()
}

// Add 添加项目目录
func (s *Service) Add(path string, label string) (domain.Project, error) {
	return s.store.AddProject(path, label)
}

// Remove 删除项目配置
func (s *Service) Remove(id string) error {
	return s.store.RemoveProject(id)
}

// Select 切换当前项目
func (s *Service) Select(id string) (domain.Project, error) {
	return s.store.SelectProject(id)
}

// SetDefault 设置默认项目
func (s *Service) SetDefault(id string) (domain.Project, error) {
	return s.store.SetDefaultProject(id)
}

// Reorder 按给定 ID 顺序重排项目列表
func (s *Service) Reorder(ids []string) ([]domain.Project, error) {
	return s.store.ReorderProjects(ids)
}
