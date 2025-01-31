package cmd

import (
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

// getAllTargetRepos merges CLI arguments, autodiscovered repos, and configured targets
func getAllTargetRepos(
	cmd *cobra.Command,
	client *gitea.Client,
	cfg *Config,
	cliArgs []string,
) ([]string, error) {
	if len(cliArgs) > 0 {
		return cliArgs, nil
	}

	var discovered []string
	if cfg.Targets.Autodiscover {
		if cfg.Targets.Organization == "" {
			return nil, fmt.Errorf("autodiscover is enabled but organization is not set in config")
		}
		filter := cfg.Targets.AutodiscoverFilter
		var err error
		discovered, err = autodiscoverRepos(client, cfg.Targets.Organization, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to autodiscover repos: %w", err)
		}
	}

	finalList := append([]string{}, discovered...)
	finalList = append(finalList, cfg.Targets.Repos...)
	finalList = deduplicate(finalList)

	if len(cfg.Targets.ExcludeRepos) > 0 {
		finalList = excludeRepos(finalList, cfg.Targets.ExcludeRepos)
	}

	return finalList, nil
}

func deduplicate(input []string) []string {
	seen := make(map[string]bool)
	var output []string
	for _, r := range input {
		lr := strings.ToLower(r)
		if !seen[lr] {
			seen[lr] = true
			output = append(output, r)
		}
	}
	return output
}

func excludeRepos(initialList, excludeList []string) []string {
	excludeMap := map[string]bool{}
	for _, e := range excludeList {
		excludeMap[strings.ToLower(e)] = true
	}

	var final []string
	for _, repo := range initialList {
		if !excludeMap[strings.ToLower(repo)] {
			final = append(final, repo)
		}
	}
	return final
}

func autodiscoverRepos(client *gitea.Client, org string, filter string) ([]string, error) {
	repos, _, err := client.ListOrgRepos(org, gitea.ListOrgReposOptions{})
	if err != nil {
		return nil, err
	}

	var results []string
	for _, r := range repos {
		if filter == "*" {
			results = append(results, fmt.Sprintf("%s/%s", org, r.Name))
		}
	}
	return results, nil
}
