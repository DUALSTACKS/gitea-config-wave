package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type TopicsConfig struct {
	Topics []string `yaml:"topics"`
}

type TopicsHandler struct{}

func (h *TopicsHandler) Name() string {
	return "topics"
}

func (h *TopicsHandler) Path() string {
	return "topics.yaml"
}

func (h *TopicsHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
	topics, _, err := client.ListRepoTopics(owner, repo, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get topics for %s/%s: %w", owner, repo, err)
	}

	return TopicsConfig{Topics: topics}, nil
}

func (h *TopicsHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
	topicsConfig, ok := data.(TopicsConfig)
	if !ok {
		return fmt.Errorf("invalid data type for TopicsHandler")
	}

	if len(topicsConfig.Topics) == 0 {
		return nil
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	strategy := cfg.TopicsUpdateStrategy
	if err := h.validateUpdateStrategy(strategy); err != nil {
		return err
	}

	if strategy == UpdateStrategyReplace {
		_, err := client.SetRepoTopics(owner, repo, topicsConfig.Topics)
		return err
	}

	for _, topic := range topicsConfig.Topics {
		_, err := client.AddRepoTopic(owner, repo, topic)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *TopicsHandler) validateUpdateStrategy(strategy UpdateStrategy) error {
	supported := map[UpdateStrategy]bool{
		UpdateStrategyReplace: true,
		UpdateStrategyAppend:  true,
	}

	if _, ok := supported[strategy]; !ok {
		return fmt.Errorf("invalid topic_update_strategy: %s (must be 'replace' or 'append')", strategy)
	}

	return nil
}

func (h *TopicsHandler) Enabled() bool {
	return true
}

func (h *TopicsHandler) Load(path string) (interface{}, error) {
	return readTopics(path)
}

func readTopics(path string) (TopicsConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return TopicsConfig{}, err
	}
	var config TopicsConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return TopicsConfig{}, err
	}
	return config, nil
}
