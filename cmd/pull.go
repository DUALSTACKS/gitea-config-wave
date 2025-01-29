package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

		log.Debug("parsing repository argument",
			"owner", owner,
			"repo", repo,
		)

		cfg, err := loadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
		fmt.Println(cfg.GiteaURL, cfg.GiteaToken)
		if err != nil {
			return fmt.Errorf("failed to create Gitea client: %w", err)
		}

		log.Debug("fetching repository settings",
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

		log.Debug("fetched repository topics",
			"owner", owner,
			"repo", repo,
			"topics", topics,
		)

		filteredRepo := toRepoSettings(repoFull, topics)

		log.Debug("fetching branch protections",
			"owner", owner,
			"repo", repo,
		)

		protections, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
		if err != nil {
			return fmt.Errorf("failed to list branch protections for %s/%s: %w", owner, repo, err)
		}

		log.Debug("fetched branch protections",
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
			outputDir = ".gitea/defaults"
		}
		if dryRun {
			log.Info("would pull repository settings (dry run)",
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

		repoSettingsPath := filepath.Join(outputDir, "repo_settings.yaml")
		if err := writeYAMLFile(repoSettingsPath, filteredRepo); err != nil {
			return fmt.Errorf("failed to write repo settings: %w", err)
		}

		branchProtectionsPath := filepath.Join(outputDir, "branch_protections.yaml")
		branchProtectionConfig := BranchProtectionConfig{
			Rules: transformedProtections,
		}
		if err := writeYAMLFile(branchProtectionsPath, branchProtectionConfig); err != nil {
			return fmt.Errorf("failed to write branch protections: %w", err)
		}

		log.Info("successfully pulled repository settings",
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

func parseRepoString(input string) (string, string, error) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.New("must be in the format owner/repo")
	}
	return parts[0], parts[1], nil
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

type RepoSettings struct {
	DefaultBranch                 string                 `json:"default_branch" yaml:"default_branch"`
	HasIssues                     bool                   `json:"has_issues" yaml:"has_issues"`
	ExternalTracker               *gitea.ExternalTracker `json:"external_tracker" yaml:"external_tracker"`
	HasWiki                       bool                   `json:"has_wiki" yaml:"has_wiki"`
	HasPullRequests               bool                   `json:"has_pull_requests" yaml:"has_pull_requests"`
	HasProjects                   bool                   `json:"has_projects" yaml:"has_projects"`
	HasReleases                   bool                   `json:"has_releases" yaml:"has_releases"`
	HasPackages                   bool                   `json:"has_packages" yaml:"has_packages"`
	HasActions                    bool                   `json:"has_actions" yaml:"has_actions"`
	IgnoreWhitespaceConflicts     bool                   `json:"ignore_whitespace_conflicts" yaml:"ignore_whitespace_conflicts"`
	AllowMergeCommits             bool                   `json:"allow_merge_commits" yaml:"allow_merge_commits"`
	AllowRebase                   bool                   `json:"allow_rebase" yaml:"allow_rebase"`
	AllowRebaseExplicit           bool                   `json:"allow_rebase_explicit" yaml:"allow_rebase_explicit"`
	AllowSquashMerge              bool                   `json:"allow_squash_merge" yaml:"allow_squash_merge"`
	DefaultDeleteBranchAfterMerge bool                   `json:"default_delete_branch_after_merge" yaml:"default_delete_branch_after_merge"`
	DefaultMergeStyle             string                 `json:"default_merge_style" yaml:"default_merge_style"`
	DefaultAllowMaintainerEdit    bool                   `json:"default_allow_maintainer_edit" yaml:"default_allow_maintainer_edit"`
	Topics                        []string               `json:"topics" yaml:"topics"`
}

type BranchProtection struct {
	BranchName                    string   `yaml:"branch_name"`
	RuleName                      string   `yaml:"rule_name"`
	EnablePush                    bool     `yaml:"enable_push"`
	EnablePushWhitelist           bool     `yaml:"enable_push_whitelist"`
	PushWhitelistUsernames        []string `yaml:"push_whitelist_usernames"`
	PushWhitelistTeams            []string `yaml:"push_whitelist_teams"`
	PushWhitelistDeployKeys       bool     `yaml:"push_whitelist_deploy_keys"`
	EnableMergeWhitelist          bool     `yaml:"enable_merge_whitelist"`
	MergeWhitelistUsernames       []string `yaml:"merge_whitelist_usernames"`
	MergeWhitelistTeams           []string `yaml:"merge_whitelist_teams"`
	EnableStatusCheck             bool     `yaml:"enable_status_check"`
	StatusCheckContexts           []string `yaml:"status_check_contexts"`
	RequiredApprovals             int64    `yaml:"required_approvals"`
	EnableApprovalsWhitelist      bool     `yaml:"enable_approvals_whitelist"`
	ApprovalsWhitelistUsernames   []string `yaml:"approvals_whitelist_usernames"`
	ApprovalsWhitelistTeams       []string `yaml:"approvals_whitelist_teams"`
	BlockOnRejectedReviews        bool     `yaml:"block_on_rejected_reviews"`
	BlockOnOfficialReviewRequests bool     `yaml:"block_on_official_review_requests"`
	BlockOnOutdatedBranch         bool     `yaml:"block_on_outdated_branch"`
	DismissStaleApprovals         bool     `yaml:"dismiss_stale_approvals"`
	RequireSignedCommits          bool     `yaml:"require_signed_commits"`
	ProtectedFilePatterns         string   `yaml:"protected_file_patterns,omitempty"`
	UnprotectedFilePatterns       string   `yaml:"unprotected_file_patterns,omitempty"`
}

type BranchProtectionConfig struct {
	Rules []BranchProtection `yaml:"rules"`
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

func writeYAMLFile(filePath string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type Config struct {
	GiteaURL   string `yaml:"gitea_url"`
	GiteaToken string `yaml:"gitea_token"`

	Config struct {
		OutputDir string `yaml:"output_dir"`
	} `yaml:"config"`

	Targets struct {
		Autodiscover       bool     `yaml:"autodiscover"`
		AutodiscoverFilter string   `yaml:"autodiscover_filter"`
		Repos              []string `yaml:"repos"`
		ExcludeRepos       []string `yaml:"exclude_repos"`
	} `yaml:"targets"`

	DryRun               bool   `yaml:"dry_run"`
	TopicsUpdateStrategy string `yaml:"topics_update_strategy"`
}

func loadConfig(filePath string) (*Config, error) {
	if filePath == "" {
		filePath = DefaultConfigFile
	}

	var cfg Config
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := godotenv.Load(); err == nil {
		if token := os.Getenv("GITEA_TOKEN"); token != "" {
			cfg.GiteaToken = token
		}
		if url := os.Getenv("GITEA_URL"); url != "" {
			cfg.GiteaURL = url
		}
	}

	return &cfg, nil
}
