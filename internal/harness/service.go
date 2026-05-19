package harness

import (
	"strings"

	"siwap/internal/config"
	"siwap/internal/domain"
)

// Service 提供助手配置的读取和变更能力
type Service struct {
	store *config.Store
}

// NewService 创建助手服务
func NewService(store *config.Store) *Service {
	return &Service{store: store}
}

// List 返回助手配置列表
func (s *Service) List() []domain.Harness {
	return s.store.ListHarnesses()
}

// Get 根据助手 ID 查找助手配置
func (s *Service) Get(id string) (domain.Harness, bool) {
	return s.store.GetHarness(id)
}

// Update 更新已有助手配置
func (s *Service) Update(next domain.Harness) (domain.Harness, error) {
	return s.store.UpdateHarness(next)
}

// Create 创建新的自定义助手
func (s *Service) Create(next domain.Harness) (domain.Harness, error) {
	return s.store.CreateHarness(next)
}

// Remove 删除指定助手
func (s *Service) Remove(id string) error {
	return s.store.RemoveHarness(id)
}

// Reorder 按给定 ID 顺序重排助手列表
func (s *Service) Reorder(ids []string) ([]domain.Harness, error) {
	return s.store.ReorderHarnesses(ids)
}

// BuildCommand 只返回用户配置的助手命令，不追加任何隐藏参数
func BuildCommand(harness domain.Harness, _ map[string]string) string {
	return strings.TrimSpace(harness.Command)
}
