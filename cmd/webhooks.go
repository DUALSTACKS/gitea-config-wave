package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type Webhook struct {
	ID     int64             `yaml:"id"`
	Type   string            `yaml:"type"`
	URL    string            `yaml:"url,omitempty"`
	Config map[string]string `yaml:"config"`
	Events []string          `yaml:"events"`
	Active bool              `yaml:"active"`
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
	config, ok := data.(WebhookConfig)
	if !ok {
		return fmt.Errorf("invalid data type for WebhooksHandler")
	}

	if len(config.Hooks) == 0 {
		return nil
	}

	// Get global config
	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// First, list existing webhooks
	existing, _, err := client.ListRepoHooks(owner, repo, gitea.ListHooksOptions{})
	if err != nil {
		return fmt.Errorf("failed to list existing webhooks: %w", err)
	}

	// If strategy is override, delete all existing webhooks
	if cfg.WebhooksUpdateStrategy == "override" {
		// Delete existing webhooks
		for _, hook := range existing {
			_, err := client.DeleteRepoHook(owner, repo, hook.ID)
			if err != nil {
				return fmt.Errorf("failed to delete webhook %d: %w", hook.ID, err)
			}
		}

		// Create new webhooks
		for _, hook := range config.Hooks {
			opt := gitea.CreateHookOption{
				Type:   gitea.HookType(hook.Type),
				Config: hook.Config,
				Events: hook.Events,
				Active: hook.Active,
			}

			_, _, err := client.CreateRepoHook(owner, repo, opt)
			if err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}
		}
		return nil
	}

	// For merge strategy, we need to:
	// 1. Keep existing webhooks that don't conflict
	// 2. Update webhooks that have matching URLs
	// 3. Add new webhooks that don't exist yet

	// Create a map of existing webhooks by URL for easy lookup
	existingByURL := make(map[string]*gitea.Hook)
	for _, hook := range existing {
		if url, ok := hook.Config["url"]; ok {
			existingByURL[url] = hook
		}
	}

	// Process each webhook in the config
	for _, hook := range config.Hooks {
		url := hook.Config["url"]
		if existing, ok := existingByURL[url]; ok {
			// Update existing webhook
			opt := gitea.EditHookOption{
				Config: hook.Config,
				Events: hook.Events,
				Active: &hook.Active,
			}
			_, err := client.EditRepoHook(owner, repo, existing.ID, opt)
			if err != nil {
				return fmt.Errorf("failed to update webhook: %w", err)
			}
			// Remove from map to mark as processed
			delete(existingByURL, url)
		} else {
			// Create new webhook
			opt := gitea.CreateHookOption{
				Type:   gitea.HookType(hook.Type),
				Config: hook.Config,
				Events: hook.Events,
				Active: hook.Active,
			}
			_, _, err := client.CreateRepoHook(owner, repo, opt)
			if err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}
		}
	}

	return nil
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
