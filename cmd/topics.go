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

	// Get global config
	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If strategy is override, just set the topics directly
	if cfg.TopicsUpdateStrategy == "override" {
		_, err := client.SetRepoTopics(owner, repo, topicsConfig.Topics)
		return err
	}

	// Otherwise merge with existing topics
	existingTopics, _, err := client.ListRepoTopics(owner, repo, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list existing topics: %w", err)
	}

	topicsMap := make(map[string]bool)
	for _, t := range existingTopics {
		topicsMap[t] = true
	}
	for _, t := range topicsConfig.Topics {
		topicsMap[t] = true
	}

	mergedTopics := make([]string, 0, len(topicsMap))
	for t := range topicsMap {
		mergedTopics = append(mergedTopics, t)
	}

	_, err = client.SetRepoTopics(owner, repo, mergedTopics)
	return err
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
