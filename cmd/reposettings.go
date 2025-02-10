package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type RepoSettingsHandler struct{}

type RepoSettings struct {
	DefaultBranch                 *string                `yaml:"default_branch,omitempty"`
	HasIssues                     *bool                  `yaml:"has_issues,omitempty"`
	ExternalTracker               *gitea.ExternalTracker `yaml:"external_tracker,omitempty"`
	HasWiki                       *bool                  `yaml:"has_wiki,omitempty"`
	HasPullRequests               *bool                  `yaml:"has_pull_requests,omitempty"`
	HasProjects                   *bool                  `yaml:"has_projects,omitempty"`
	HasReleases                   *bool                  `yaml:"has_releases,omitempty"`
	HasPackages                   *bool                  `yaml:"has_packages,omitempty"`
	HasActions                    *bool                  `yaml:"has_actions,omitempty"`
	IgnoreWhitespaceConflicts     *bool                  `yaml:"ignore_whitespace_conflicts,omitempty"`
	AllowMergeCommits             *bool                  `yaml:"allow_merge_commits,omitempty"`
	AllowRebase                   *bool                  `yaml:"allow_rebase,omitempty"`
	AllowRebaseExplicit           *bool                  `yaml:"allow_rebase_explicit,omitempty"`
	AllowSquashMerge              *bool                  `yaml:"allow_squash_merge,omitempty"`
	DefaultDeleteBranchAfterMerge *bool                  `yaml:"default_delete_branch_after_merge,omitempty"`
	DefaultMergeStyle             *string                `yaml:"default_merge_style,omitempty"`
	DefaultAllowMaintainerEdit    *bool                  `yaml:"default_allow_maintainer_edit,omitempty"`
	Topics                        []string               `yaml:"topics,omitempty"`
}

func (h *RepoSettingsHandler) Name() string {
	return "repository settings"
}

func (h *RepoSettingsHandler) Path() string {
	return DefaultRepoSettingsFile
}

func (h *RepoSettingsHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
	repoSDK, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo %s/%s: %w", owner, repo, err)
	}

	return toRepoSettings(repoSDK), nil
}

func (h *RepoSettingsHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
	rs, ok := data.(*RepoSettings)
	if !ok {
		return fmt.Errorf("invalid data type for RepoSettingsHandler")
	}

	// Update repository settings
	_, _, err := client.EditRepo(owner, repo, toEditRepoOption(rs))
	return err
}

func toRepoSettings(gr *gitea.Repository) *RepoSettings {
	defaultBranch := gr.DefaultBranch
	hasIssues := gr.HasIssues
	hasWiki := gr.HasWiki
	hasPullRequests := gr.HasPullRequests
	hasProjects := gr.HasProjects
	hasReleases := gr.HasReleases
	hasPackages := gr.HasPackages
	hasActions := gr.HasActions
	ignoreWhitespaceConflicts := gr.IgnoreWhitespaceConflicts
	allowMergeCommits := gr.AllowMerge
	allowRebase := gr.AllowRebase
	allowRebaseExplicit := gr.AllowRebaseMerge
	allowSquashMerge := gr.AllowSquash
	defaultDeleteBranchAfterMerge := false
	defaultMergeStyle := string(gr.DefaultMergeStyle)
	defaultAllowMaintainerEdit := false

	return &RepoSettings{
		DefaultBranch:                 &defaultBranch,
		HasIssues:                     &hasIssues,
		ExternalTracker:               gr.ExternalTracker,
		HasWiki:                       &hasWiki,
		HasPullRequests:               &hasPullRequests,
		HasProjects:                   &hasProjects,
		HasReleases:                   &hasReleases,
		HasPackages:                   &hasPackages,
		HasActions:                    &hasActions,
		IgnoreWhitespaceConflicts:     &ignoreWhitespaceConflicts,
		AllowMergeCommits:             &allowMergeCommits,
		AllowRebase:                   &allowRebase,
		AllowRebaseExplicit:           &allowRebaseExplicit,
		AllowSquashMerge:              &allowSquashMerge,
		DefaultDeleteBranchAfterMerge: &defaultDeleteBranchAfterMerge,
		DefaultMergeStyle:             &defaultMergeStyle,
		DefaultAllowMaintainerEdit:    &defaultAllowMaintainerEdit,
	}
}

func toEditRepoOption(rs *RepoSettings) gitea.EditRepoOption {
	return gitea.EditRepoOption{
		DefaultBranch:             rs.DefaultBranch,
		HasIssues:                 rs.HasIssues,
		ExternalTracker:           rs.ExternalTracker,
		HasWiki:                   rs.HasWiki,
		HasPullRequests:           rs.HasPullRequests,
		HasProjects:               rs.HasProjects,
		HasReleases:               rs.HasReleases,
		HasPackages:               rs.HasPackages,
		HasActions:                rs.HasActions,
		IgnoreWhitespaceConflicts: rs.IgnoreWhitespaceConflicts,
		AllowMerge:                rs.AllowMergeCommits,
		AllowRebase:               rs.AllowRebase,
		AllowRebaseMerge:          rs.AllowRebaseExplicit,
		AllowSquash:               rs.AllowSquashMerge,
	}
}

func (h *RepoSettingsHandler) Enabled() bool {
	return true
}

func (h *RepoSettingsHandler) Load(path string) (interface{}, error) {
	return readRepoSettings(path)
}

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
