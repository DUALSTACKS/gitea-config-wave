package cmd

// TODO: webhooks can support replace, append and merge (by target url but only the first webhook with this target url will be "merged")

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type Webhook struct {
	ID                  int64             `yaml:"id"`
	Type                string            `yaml:"type"`
	URL                 string            `yaml:"url,omitempty"`
	BranchFilter        string            `yaml:"branch_filter,omitempty"`
	Config              map[string]string `yaml:"config"`
	Events              []string          `yaml:"events"`
	Active              bool              `yaml:"active"`
	AuthorizationHeader string            `yaml:"authorization_header,omitempty"`
}

type WebhookConfig struct {
	Hooks []Webhook `yaml:"hooks"`
}

type WebhooksHandler struct{}

func (h *WebhooksHandler) Name() string {
	return "webhooks"
}

func (h *WebhooksHandler) Path() string {
	return DefaultWebhooksFile
}

func (h *WebhooksHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
	webhooks, _, err := client.ListRepoHooks(owner, repo, gitea.ListHooksOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks for %s/%s: %w", owner, repo, err)
	}

	transformed := make([]Webhook, len(webhooks))
	for i, wh := range webhooks {
		transformed[i] = Webhook{
			ID:     wh.ID,
			Type:   wh.Type,
			URL:    wh.URL,
			Config: wh.Config,
			Events: wh.Events,
			Active: wh.Active,
		}
	}
	return WebhookConfig{Hooks: transformed}, nil
}

func (h *WebhooksHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
	whConfig, ok := data.(WebhookConfig)
	if !ok {
		return fmt.Errorf("invalid data type for WebhooksHandler")
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	strategy := cfg.WebhooksUpdateStrategy
	if err := h.validateUpdateStrategy(strategy); err != nil {
		return err
	}

	existingByID, countByURL, err := h.getExistingWebhooksMap(client, owner, repo)
	if err != nil {
		return err
	}

	if strategy == UpdateStrategyAppend {
		for _, wh := range whConfig.Hooks {
			// skip if there's already a webhook with the same target URL
			if _, ok := countByURL[wh.URL]; ok {
				continue
			}

			_, _, err := client.CreateRepoHook(owner, repo, toCreateHookOption(wh))
			if err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}
		}

		return nil
	}

	if strategy == UpdateStrategyReplace {
		for _, wh := range existingByID {
			_, err := client.DeleteRepoHook(owner, repo, wh.ID)
			if err != nil {
				return fmt.Errorf("failed to delete webhook: %w", err)
			}
			delete(existingByID, wh.ID)
		}
	}

	for _, wh := range whConfig.Hooks {
		if count, ok := countByURL[wh.URL]; ok {
			if count > 1 {
				// skipping because there are multiple webhooks with the same target URL
				continue
			}

			_, err := client.EditRepoHook(owner, repo, wh.ID, toEditHookOption(wh))
			if err != nil {
				return fmt.Errorf("failed to update webhook: %w", err)
			}

			continue
		}

		_, _, err := client.CreateRepoHook(owner, repo, toCreateHookOption(wh))
		if err != nil {
			return fmt.Errorf("failed to create webhook: %w", err)
		}

	}

	return nil
}

func toEditHookOption(wh Webhook) gitea.EditHookOption {
	return gitea.EditHookOption{
		Config:              wh.Config,
		Events:              wh.Events,
		BranchFilter:        wh.BranchFilter,
		Active:              &wh.Active,
		AuthorizationHeader: wh.AuthorizationHeader, // TODO: how to deal with secrets?
	}
}

func toCreateHookOption(wh Webhook) gitea.CreateHookOption {
	return gitea.CreateHookOption{
		Type: gitea.HookType(wh.Type),
		// URL:          wh.URL, // TODO: not supported yet
		// Method:       wh.Method, // TODO: for some reason not returned by the Gitea API
		Config:              wh.Config,
		Events:              wh.Events,
		BranchFilter:        wh.BranchFilter,
		Active:              wh.Active,
		AuthorizationHeader: wh.AuthorizationHeader, // TODO: how to deal with secrets?
	}
}

func (h *WebhooksHandler) Enabled() bool {
	return true
}

func (h *WebhooksHandler) Load(path string) (interface{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config WebhookConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func (h *WebhooksHandler) getExistingWebhooksMap(client *gitea.Client, owner, repo string) (map[int64]*gitea.Hook, map[string]int, error) {
	webhooks, _, err := client.ListRepoHooks(owner, repo, gitea.ListHooksOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	byID := make(map[int64]*gitea.Hook, len(webhooks))
	countByURL := make(map[string]int, len(webhooks))

	for _, wh := range webhooks {
		byID[wh.ID] = wh
		countByURL[wh.URL]++
	}

	return byID, countByURL, nil
}

func (h *WebhooksHandler) validateUpdateStrategy(strategy UpdateStrategy) error {
	supported := map[UpdateStrategy]bool{
		UpdateStrategyReplace: true,
		UpdateStrategyMerge:   true,
		UpdateStrategyAppend:  true,
	}

	if _, ok := supported[strategy]; !ok {
		return fmt.Errorf("invalid webhooks_update_strategy: %s (must be 'replace', 'merge', or 'append')", strategy)
	}

	return nil
}
