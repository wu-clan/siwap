package harness

import (
	"strings"

	"siwap/internal/config"
	"siwap/internal/domain"
)

type Service struct {
	store *config.Store
}

func NewService(store *config.Store) *Service {
	return &Service{store: store}
}

func (s *Service) List() []domain.Harness {
	return s.store.ListHarnesses()
}

func (s *Service) Get(id string) (domain.Harness, bool) {
	return s.store.GetHarness(id)
}

func (s *Service) Update(next domain.Harness) (domain.Harness, error) {
	return s.store.UpdateHarness(next)
}

func (s *Service) Create(next domain.Harness) (domain.Harness, error) {
	return s.store.CreateHarness(next)
}

func (s *Service) Remove(id string) error {
	return s.store.RemoveHarness(id)
}

func (s *Service) Reorder(ids []string) ([]domain.Harness, error) {
	return s.store.ReorderHarnesses(ids)
}

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

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
