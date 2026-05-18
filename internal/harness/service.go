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

// BuildCommand 根据助手配置和参数值生成启动命令
func BuildCommand(harness domain.Harness, overrides map[string]string) string {
	flags := map[string]string{}
	for key, value := range harness.Flags {
		flags[key] = value
	}
	for key, value := range overrides {
		flags[key] = value
	}

	parts := []string{strings.TrimSpace(harness.Command)}
	for _, option := range harness.FlagOptions {
		value := flags[option.Key]
		if value == "" {
			value = option.Default
		}
		switch option.Type {
		case "toggle":
			if value == "true" && option.CommandFlag != "" {
				parts = append(parts, option.CommandFlag)
			}
		case "select":
			if value != "" && value != option.Default && option.CommandFlag != "" {
				parts = append(parts, option.CommandFlag, shellQuote(value))
			}
		}
	}
	return strings.Join(parts, " ")
}

// shellQuote 对命令参数进行 shell 安全转义
func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
