package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// pushCmd handles pushing local repository settings to Gitea instances
var pushCmd = &cobra.Command{
	Use:   "push [owner/repo]...",
	Short: "Push repository settings to a Gitea instance",
	Long: `Pushes (applies) the local repository settings (e.g., branch 
protections, topics, etc.) to one or more Gitea repositories specified 
in the config file or in the command arguments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("could not parse --dry-run flag: %w", err)
		}

		cfg, err := loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
		if err != nil {
			return fmt.Errorf("failed to create Gitea client: %w", err)
		}

		targetRepos, err := computeTargetRepos(cmd, client, cfg, args)
		if err != nil {
			return err
		}
		if len(targetRepos) == 0 {
			return errors.New("no repositories to process after merges/exclusions")
		}

		log.Info("computed target repositories", "count", len(targetRepos), "repos", targetRepos)

		outputDir := cfg.Config.OutputDir
		if outputDir == "" {
			outputDir = ".gitea/defaults"
		}

		repoSettingsPath := filepath.Join(outputDir, "repo_settings.yaml")
		branchProtectionsPath := filepath.Join(outputDir, "branch_protections.yaml")

		localRepoSettings, err := readRepoSettings(repoSettingsPath)
		if err != nil {
			return fmt.Errorf("failed to read local repo settings: %w", err)
		}

		localBranchProtections, err := readBranchProtections(branchProtectionsPath)
		if err != nil {
			return fmt.Errorf("failed to read local branch protections: %w", err)
		}

		for _, fullName := range targetRepos {
			owner, repo, err := parseRepoString(fullName)
			if err != nil {
				return fmt.Errorf("invalid repo argument %q: %w", fullName, err)
			}

			if dryRun || cfg.DryRun {
				log.Info("would push settings (dry run)",
					"owner", owner,
					"repo", repo,
				)
				continue
			}

			log.Debug("updating repository settings",
				"owner", owner,
				"repo", repo,
			)

			editOpt := toEditRepoOption(localRepoSettings)
			_, resp, err := client.EditRepo(owner, repo, gitea.EditRepoOption(editOpt))
			if err != nil {
				return fmt.Errorf("failed to edit repo %s/%s: %w", owner, repo, err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status %d editing repo %s/%s", resp.StatusCode, owner, repo)
			}

			log.Debug("updating repository topics",
				"owner", owner,
				"repo", repo,
				"topics", localRepoSettings.Topics,
			)

			err = mergeTopics(client, owner, repo, localRepoSettings.Topics, cfg)
			if err != nil {
				return fmt.Errorf("failed to update topics for %s/%s: %w", owner, repo, err)
			}

			existingProtections, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
			if err != nil {
				return fmt.Errorf("failed to list existing branch protections for %s/%s: %w", owner, repo, err)
			}

			log.Debug("removing existing branch protections",
				"owner", owner,
				"repo", repo,
				"count", len(existingProtections),
			)

			for _, bp := range existingProtections {
				_, err := client.DeleteBranchProtection(owner, repo, bp.BranchName)
				if err != nil {
					return fmt.Errorf("failed to delete branch protection for %s/%s branch %s: %w",
						owner, repo, bp.BranchName, err)
				}
			}

			log.Debug("creating new branch protections",
				"owner", owner,
				"repo", repo,
				"count", len(localBranchProtections),
			)

			for _, bp := range localBranchProtections {
				bpOpt := gitea.CreateBranchProtectionOption{
					BranchName:                    bp.BranchName,
					RuleName:                      bp.RuleName,
					EnablePush:                    bp.EnablePush,
					EnablePushWhitelist:           bp.EnablePushWhitelist,
					PushWhitelistUsernames:        bp.PushWhitelistUsernames,
					PushWhitelistTeams:            bp.PushWhitelistTeams,
					PushWhitelistDeployKeys:       bp.PushWhitelistDeployKeys,
					EnableMergeWhitelist:          bp.EnableMergeWhitelist,
					MergeWhitelistUsernames:       bp.MergeWhitelistUsernames,
					MergeWhitelistTeams:           bp.MergeWhitelistTeams,
					EnableStatusCheck:             bp.EnableStatusCheck,
					StatusCheckContexts:           bp.StatusCheckContexts,
					RequiredApprovals:             bp.RequiredApprovals,
					EnableApprovalsWhitelist:      bp.EnableApprovalsWhitelist,
					ApprovalsWhitelistUsernames:   bp.ApprovalsWhitelistUsernames,
					ApprovalsWhitelistTeams:       bp.ApprovalsWhitelistTeams,
					BlockOnRejectedReviews:        bp.BlockOnRejectedReviews,
					BlockOnOfficialReviewRequests: bp.BlockOnOfficialReviewRequests,
					BlockOnOutdatedBranch:         bp.BlockOnOutdatedBranch,
					DismissStaleApprovals:         bp.DismissStaleApprovals,
					RequireSignedCommits:          bp.RequireSignedCommits,
					ProtectedFilePatterns:         bp.ProtectedFilePatterns,
					UnprotectedFilePatterns:       bp.UnprotectedFilePatterns,
				}
				_, _, err = client.CreateBranchProtection(owner, repo, bpOpt)
				if err != nil {
					return fmt.Errorf("failed to create branch protection for %s/%s: %w", owner, repo, err)
				}
			}

			log.Info("successfully pushed settings",
				"owner", owner,
				"repo", repo,
				"topics_count", len(localRepoSettings.Topics),
				"protections_count", len(localBranchProtections),
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

// computeTargetRepos merges CLI arguments, autodiscovered repos, and configured targets
func computeTargetRepos(
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
		orgName := "DUALSTACKS"
		filter := cfg.Targets.AutodiscoverFilter
		var err error
		discovered, err = autodiscoverRepos(client, orgName, filter)
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

// readRepoSettings loads the minimal RepoSettings YAML from disk
func readRepoSettings(path string) (*RepoSettings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rs RepoSettings
	if err := yaml.Unmarshal(b, &rs); err != nil {
		return nil, err
	}
	return &rs, nil
}

// readBranchProtections loads an array of BranchProtection objects from disk
func readBranchProtections(path string) ([]BranchProtection, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bpConfig BranchProtectionConfig
	if err := yaml.Unmarshal(b, &bpConfig); err != nil {
		return nil, err
	}
	return bpConfig.Rules, nil
}

// toEditRepoOption transforms your minimal RepoSettings struct into gitea.EditRepoOption
func toEditRepoOption(rs *RepoSettings) gitea.EditRepoOption {
	return gitea.EditRepoOption{
		DefaultBranch:             &rs.DefaultBranch,
		HasIssues:                 &rs.HasIssues,
		ExternalTracker:           rs.ExternalTracker,
		HasWiki:                   &rs.HasWiki,
		HasPullRequests:           &rs.HasPullRequests,
		HasProjects:               &rs.HasProjects,
		IgnoreWhitespaceConflicts: &rs.IgnoreWhitespaceConflicts,
		AllowMerge:                &rs.AllowMergeCommits,
		AllowRebase:               &rs.AllowRebase,
		AllowRebaseMerge:          &rs.AllowRebaseExplicit,
		AllowSquash:               &rs.AllowSquashMerge,
		HasReleases:               &rs.HasReleases,
		HasActions:                &rs.HasActions,
		// The rest fields can be filled if supported by your Gitea version
	}
}

// mergeTopics updates the repo topics according to the configured strategy (merge or override)
func mergeTopics(client *gitea.Client, owner, repo string, localTopics []string, cfg *Config) error {
	if len(localTopics) == 0 {
		// nothing to do
		return nil
	}

	// If strategy is override, just set the topics directly
	if cfg.TopicsUpdateStrategy == "override" {
		_, err := client.SetRepoTopics(owner, repo, localTopics)
		return err
	}

	// Otherwise use merge strategy (default behavior)
	existingTopics, _, err := client.ListRepoTopics(owner, repo, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list existing topics: %w", err)
	}

	topicsMap := make(map[string]bool)
	for _, t := range existingTopics {
		topicsMap[t] = true
	}
	for _, t := range localTopics {
		topicsMap[t] = true
	}

	mergedTopics := make([]string, 0, len(topicsMap))
	for t := range topicsMap {
		mergedTopics = append(mergedTopics, t)
	}

	_, err = client.SetRepoTopics(owner, repo, mergedTopics)
	return err
}
