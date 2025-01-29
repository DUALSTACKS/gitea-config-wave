package cmd

import "code.gitea.io/sdk/gitea"

type Config struct {
	GiteaURL   string `yaml:"gitea_url" validate:"required,url"`
	GiteaToken string `yaml:"gitea_token" validate:"required"`
	Config     struct {
		OutputDir string `yaml:"output_dir" validate:"omitempty,dirpath"`
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

type RepoSettings struct {
	DefaultBranch                 string                 `yaml:"default_branch"`
	HasIssues                     bool                   `yaml:"has_issues"`
	ExternalTracker               *gitea.ExternalTracker `yaml:"external_tracker"`
	HasWiki                       bool                   `yaml:"has_wiki"`
	HasPullRequests               bool                   `yaml:"has_pull_requests"`
	HasProjects                   bool                   `yaml:"has_projects"`
	HasReleases                   bool                   `yaml:"has_releases"`
	HasPackages                   bool                   `yaml:"has_packages"`
	HasActions                    bool                   `yaml:"has_actions"`
	IgnoreWhitespaceConflicts     bool                   `yaml:"ignore_whitespace_conflicts"`
	AllowMergeCommits             bool                   `yaml:"allow_merge_commits"`
	AllowRebase                   bool                   `yaml:"allow_rebase"`
	AllowRebaseExplicit           bool                   `yaml:"allow_rebase_explicit"`
	AllowSquashMerge              bool                   `yaml:"allow_squash_merge"`
	DefaultDeleteBranchAfterMerge bool                   `yaml:"default_delete_branch_after_merge"`
	DefaultMergeStyle             string                 `yaml:"default_merge_style"`
	DefaultAllowMaintainerEdit    bool                   `yaml:"default_allow_maintainer_edit"`
	Topics                        []string               `yaml:"topics"`
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
