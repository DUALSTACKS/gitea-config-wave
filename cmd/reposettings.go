package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type RepoSettingsHandler struct{}

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
	}
}

func toEditRepoOption(rs *RepoSettings) gitea.EditRepoOption {
	defaultBranch := rs.DefaultBranch // Create a copy to get address of
	return gitea.EditRepoOption{
		DefaultBranch:             &defaultBranch,
		HasIssues:                 &rs.HasIssues,
		ExternalTracker:           rs.ExternalTracker,
		HasWiki:                   &rs.HasWiki,
		HasPullRequests:           &rs.HasPullRequests,
		HasProjects:               &rs.HasProjects,
		HasReleases:               &rs.HasReleases,
		HasPackages:               &rs.HasPackages,
		HasActions:                &rs.HasActions,
		IgnoreWhitespaceConflicts: &rs.IgnoreWhitespaceConflicts,
		AllowMerge:                &rs.AllowMergeCommits,
		AllowRebase:               &rs.AllowRebase,
		AllowRebaseMerge:          &rs.AllowRebaseExplicit,
		AllowSquash:               &rs.AllowSquashMerge,
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
