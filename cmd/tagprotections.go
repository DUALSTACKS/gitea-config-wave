package cmd

// TODO: Not supported yet (Go Gitea SDK does not support Tag Protections)

// import (
// 	"fmt"
// 	"os"

// 	"gopkg.in/yaml.v3"
// )

// type TagProtectionsHandler struct{}

// type TagProtection struct {
// 	NamePattern        string `yaml:"name_pattern"`
// 	WhitelistUsernames string `yaml:"whitelist_usernames"`
// 	WhitelistTeams     string `yaml:"whitelist_teams"`
// }

// type TagProtectionConfig struct {
// 	Rules []TagProtection `yaml:"rules"`
// }

// func (h *TagProtectionsHandler) Name() string {
// 	return "tag protections"
// }

// func (h *TagProtectionsHandler) Path() string {
// 	return DefaultTagProtectionsFile
// }

// func (h *TagProtectionsHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
// 	protections, _, err := client.ListTagProtection(owner, repo, gitea.ListRepoTagProtectionsOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list tag protections for %s/%s: %w", owner, repo, err)
// 	}

// 	transformedProtections := make([]TagProtection, len(protections))
// 	for i, tp := range protections {
// 		transformedProtections[i] = TagProtection{
// 			NamePattern:        tp.NamePattern,
// 			WhitelistUsernames: tp.WhitelistUsernames,
// 			WhitelistTeams:     tp.WhitelistTeams,
// 		}
// 	}

// 	return TagProtectionConfig{Rules: transformedProtections}, nil
// }

// func (h *TagProtectionsHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
// 	tpConfig, ok := data.(TagProtectionConfig)
// 	if !ok {
// 		return fmt.Errorf("invalid data type for TagProtectionsHandler")
// 	}

// 	cfg, err := LoadConfig(cfgFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to load config: %w", err)
// 	}

// 	strategy := cfg.TagProtectionsUpdateStrategy
// 	if err := h.validateUpdateStrategy(strategy); err != nil {
// 		return err
// 	}

// 	existing, err := h.getExistingProtectionsMap(client, owner, repo)
// 	if err != nil {
// 		return err
// 	}

// 	if strategy == UpdateStrategyAppend {
// 		for _, tp := range tpConfig.Rules {
// 			if _, ok := existing[tp.NamePattern]; ok {
// 				continue
// 			}

// 			_, _, err := client.CreateTagProtection(owner, repo, gitea.CreateTagProtectionOption{
// 				NamePattern:        tp.NamePattern,
// 				WhitelistUsernames: tp.WhitelistUsernames,
// 				WhitelistTeams:     tp.WhitelistTeams,
// 			})
// 			if err != nil {
// 				return fmt.Errorf("failed to create protection: %w", err)
// 			}
// 		}
// 		return nil
// 	}

// 	if strategy == UpdateStrategyReplace {
// 		for _, tp := range existing {
// 			_, err := client.DeleteTagProtection(owner, repo, tp.ID)
// 			if err != nil {
// 				return fmt.Errorf("failed to delete tag protection: %w", err)
// 			}
// 			delete(existing, tp.NamePattern)
// 		}
// 	}

// 	for _, tp := range tpConfig.Rules {
// 		_, _, err := client.CreateTagProtection(owner, repo, gitea.CreateTagProtectionOption{
// 			NamePattern:        tp.NamePattern,
// 			WhitelistUsernames: tp.WhitelistUsernames,
// 			WhitelistTeams:     tp.WhitelistTeams,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("failed to create tag protection: %w", err)
// 		}
// 	}

// 	return nil
// }

// func (h *TagProtectionsHandler) getExistingProtectionsMap(client *gitea.Client, owner, repo string) (map[string]*gitea.TagProtection, error) {
// 	protections, _, err := client.ListTagProtections(owner, repo, gitea.ListRepoTagProtectionsOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list tag protections: %w", err)
// 	}

// 	m := make(map[string]*gitea.TagProtection, len(protections))
// 	for _, tp := range protections {
// 		m[tp.NamePattern] = tp
// 	}

// 	return m, nil
// }

// func (h *TagProtectionsHandler) validateUpdateStrategy(strategy UpdateStrategy) error {
// 	supported := map[UpdateStrategy]bool{
// 		UpdateStrategyReplace: true,
// 		UpdateStrategyMerge:   true,
// 		UpdateStrategyAppend:  true,
// 	}

// 	if _, ok := supported[strategy]; !ok {
// 		return fmt.Errorf("invalid tag_protections_update_strategy: %s (must be 'replace', 'merge', or 'append')", strategy)
// 	}

// 	return nil
// }

// func (h *TagProtectionsHandler) Enabled() bool {
// 	return true
// }

// func readTagProtections(path string) (TagProtectionConfig, error) {
// 	b, err := os.ReadFile(path)
// 	if err != nil {
// 		return TagProtectionConfig{}, err
// 	}
// 	var tpConfig TagProtectionConfig
// 	if err := yaml.Unmarshal(b, &tpConfig); err != nil {
// 		return TagProtectionConfig{}, err
// 	}
// 	return tpConfig, nil
// }

// func (h *TagProtectionsHandler) Load(path string) (interface{}, error) {
// 	return readTagProtections(path)
// }
