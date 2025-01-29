package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

// pullCmd handles pulling repository settings from Gitea instances
var pullCmd = &cobra.Command{
	Use:   "pull [owner/repo]",
	Short: "Pull repository settings from a Gitea instance",
	Long: `Pulls the repository settings (e.g., branch protections,
issues/PR templates, etc.) from a specified Gitea repository and 
stores them locally as YAML files (by default in .gitea/defaults).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("could not parse --dry-run flag: %w", err)
		}

		owner, repo, err := parseRepoString(args[0])
		if err != nil {
			return fmt.Errorf("invalid repo argument %q: %w", args[0], err)
		}

		logger.Debug("parsing repository argument",
			"owner", owner,
			"repo", repo,
		)

		cfg, err := LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := GiteaClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Gitea client: %w", err)
		}

		logger.Debug("fetching repository settings",
			"owner", owner,
			"repo", repo,
		)

		repoFull, topics, resp, err := getRepoAndTopics(client, owner, repo)
		if err != nil {
			return fmt.Errorf("failed to get repository %s/%s (with topics): %w", owner, repo, err)
		}
		if resp != nil && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code %d while getting repo", resp.StatusCode)
		}

		logger.Debug("fetched repository topics",
			"owner", owner,
			"repo", repo,
			"topics", topics,
		)

		filteredRepo := toRepoSettings(repoFull, topics)

		logger.Debug("fetching branch protections",
			"owner", owner,
			"repo", repo,
		)

		protections, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
		if err != nil {
			return fmt.Errorf("failed to list branch protections for %s/%s: %w", owner, repo, err)
		}

		logger.Debug("fetched branch protections",
			"owner", owner,
			"repo", repo,
			"count", len(protections),
		)

		transformedProtections := make([]BranchProtection, len(protections))
		for i, bp := range protections {
			transformedProtections[i] = BranchProtection{
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
		}

		outputDir := cfg.Config.OutputDir
		if outputDir == "" {
			outputDir = DefaultOutputDir
		}
		if dryRun {
			logger.Info("would pull repository settings (dry run)",
				"owner", owner,
				"repo", repo,
				"output_dir", outputDir,
				"topics_count", len(topics),
				"protections_count", len(protections),
			)
			return nil
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
		}

		repoSettingsPath := filepath.Join(outputDir, DefaultRepoSettingsFile)
		if err := WriteYAMLFile(repoSettingsPath, filteredRepo); err != nil {
			return fmt.Errorf("failed to write repo settings: %w", err)
		}

		branchProtectionsPath := filepath.Join(outputDir, DefaultBranchProtectionsFile)
		branchProtectionConfig := BranchProtectionConfig{
			Rules: transformedProtections,
		}
		if err := WriteYAMLFile(branchProtectionsPath, branchProtectionConfig); err != nil {
			return fmt.Errorf("failed to write branch protections: %w", err)
		}

		logger.Info("successfully pulled repository settings",
			"owner", owner,
			"repo", repo,
			"output_dir", outputDir,
			"topics_count", len(topics),
			"protections_count", len(protections),
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

// RepoWithTopics extends gitea.Repository to include topics
type RepoWithTopics struct {
	gitea.Repository
	Topics []string `json:"topics"`
}

// getRepoAndTopics fetches repository info and topics in a single request
func getRepoAndTopics(client *gitea.Client, owner, repo string) (
	repoSDK *gitea.Repository,
	topics []string,
	resp *gitea.Response,
	err error,
) {
	repoSDK, resp, err = client.GetRepo(owner, repo)
	if err != nil {
		return nil, nil, resp, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}

	topics, resp, err = client.ListRepoTopics(owner, repo, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return repoSDK, nil, resp, fmt.Errorf("failed to get topics for %s/%s: %w", owner, repo, err)
	}

	return repoSDK, topics, resp, nil
}

func toRepoSettings(gr *gitea.Repository, topics []string) *RepoSettings {
	return &RepoSettings{
		DefaultBranch:                 gr.DefaultBranch,
		HasIssues:                     gr.HasIssues,
		ExternalTracker:               gr.ExternalTracker,
		HasWiki:                       gr.HasWiki,
		HasPullRequests:               gr.HasPullRequests,
		HasProjects:                   gr.HasProjects,
		HasReleases:                   gr.HasReleases,
		HasPackages:                   gr.HasPackages,
		HasActions:                    gr.HasActions,
		IgnoreWhitespaceConflicts:     gr.IgnoreWhitespaceConflicts,
		AllowMergeCommits:             gr.AllowMerge,
		AllowRebase:                   gr.AllowRebase,
		AllowRebaseExplicit:           gr.AllowRebaseMerge,
		AllowSquashMerge:              gr.AllowSquash,
		DefaultDeleteBranchAfterMerge: false,
		DefaultMergeStyle:             string(gr.DefaultMergeStyle),
		DefaultAllowMaintainerEdit:    false,
		Topics:                        topics,
	}
}
