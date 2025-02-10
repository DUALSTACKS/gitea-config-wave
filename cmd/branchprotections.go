package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type BranchProtectionsHandler struct{}

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

func (h *BranchProtectionsHandler) Name() string {
	return "branch protections"
}

func (h *BranchProtectionsHandler) Path() string {
	return DefaultBranchProtectionsFile
}

func (h *BranchProtectionsHandler) Pull(client *gitea.Client, owner, repo string) (interface{}, error) {
	protections, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list branch protections for %s/%s: %w", owner, repo, err)
	}

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

	return BranchProtectionConfig{Rules: transformedProtections}, nil
}

func (h *BranchProtectionsHandler) Push(client *gitea.Client, owner, repo string, data interface{}) error {
	bpConfig, ok := data.(BranchProtectionConfig)
	if !ok {
		return fmt.Errorf("invalid data type for BranchProtectionsHandler")
	}

	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	strategy := cfg.BranchProtectionsUpdateStrategy
	if err := h.validateUpdateStrategy(strategy); err != nil {
		return err
	}

	existing, err := h.getExistingProtectionsMap(client, owner, repo)
	if err != nil {
		return err
	}

	if strategy == UpdateStrategyAppend {
		for _, bp := range bpConfig.Rules {
			if _, ok := existing[bp.RuleName]; ok {
				continue
			}

			_, _, err := client.CreateBranchProtection(owner, repo, toCreateBranchProtectionOption(bp))
			if err != nil {
				return fmt.Errorf("failed to create protection: %w", err)
			}
		}

		return nil
	}

	if strategy == UpdateStrategyReplace {
		for _, bp := range existing {
			_, err := client.DeleteBranchProtection(owner, repo, bp.RuleName)
			if err != nil {
				return fmt.Errorf("failed to delete branch protection: %w", err)
			}
			delete(existing, bp.RuleName)
		}
	}

	for _, bp := range bpConfig.Rules {
		if _, ok := existing[bp.BranchName]; ok {
			_, _, err := client.EditBranchProtection(owner, repo, bp.RuleName, toEditBranchProtectionOption(bp))
			if err != nil {
				return fmt.Errorf("failed to update branch protection: %w", err)
			}
			continue
		}

		_, _, err := client.CreateBranchProtection(owner, repo, toCreateBranchProtectionOption(bp))
		if err != nil {
			return fmt.Errorf("failed to create branch protection: %w", err)
		}

	}

	return nil
}

func toCreateBranchProtectionOption(bp BranchProtection) gitea.CreateBranchProtectionOption {
	return gitea.CreateBranchProtectionOption{
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

func toEditBranchProtectionOption(bp BranchProtection) gitea.EditBranchProtectionOption {
	return gitea.EditBranchProtectionOption{
		EnablePush:                    &bp.EnablePush,
		EnablePushWhitelist:           &bp.EnablePushWhitelist,
		PushWhitelistUsernames:        bp.PushWhitelistUsernames,
		PushWhitelistTeams:            bp.PushWhitelistTeams,
		PushWhitelistDeployKeys:       &bp.PushWhitelistDeployKeys,
		EnableMergeWhitelist:          &bp.EnableMergeWhitelist,
		MergeWhitelistUsernames:       bp.MergeWhitelistUsernames,
		MergeWhitelistTeams:           bp.MergeWhitelistTeams,
		EnableStatusCheck:             &bp.EnableStatusCheck,
		StatusCheckContexts:           bp.StatusCheckContexts,
		RequiredApprovals:             &bp.RequiredApprovals,
		EnableApprovalsWhitelist:      &bp.EnableApprovalsWhitelist,
		ApprovalsWhitelistUsernames:   bp.ApprovalsWhitelistUsernames,
		ApprovalsWhitelistTeams:       bp.ApprovalsWhitelistTeams,
		BlockOnRejectedReviews:        &bp.BlockOnRejectedReviews,
		BlockOnOfficialReviewRequests: &bp.BlockOnOfficialReviewRequests,
		BlockOnOutdatedBranch:         &bp.BlockOnOutdatedBranch,
		DismissStaleApprovals:         &bp.DismissStaleApprovals,
		RequireSignedCommits:          &bp.RequireSignedCommits,
		ProtectedFilePatterns:         &bp.ProtectedFilePatterns,
		UnprotectedFilePatterns:       &bp.UnprotectedFilePatterns,
	}
}

func (h *BranchProtectionsHandler) Enabled() bool {
	return true
}

func (h *BranchProtectionsHandler) getExistingProtectionsMap(client *gitea.Client, owner, repo string) (map[string]*gitea.BranchProtection, error) {
	protections, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list branch protections: %w", err)
	}

	m := make(map[string]*gitea.BranchProtection, len(protections))
	for _, bp := range protections {
		m[bp.RuleName] = bp
	}

	return m, nil
}

func readBranchProtections(path string) (BranchProtectionConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return BranchProtectionConfig{}, err
	}
	var bpConfig BranchProtectionConfig
	if err := yaml.Unmarshal(b, &bpConfig); err != nil {
		return BranchProtectionConfig{}, err
	}
	return bpConfig, nil
}

func (h *BranchProtectionsHandler) Load(path string) (interface{}, error) {
	return readBranchProtections(path)
}

func (h *BranchProtectionsHandler) validateUpdateStrategy(strategy UpdateStrategy) error {
	supported := map[UpdateStrategy]bool{
		UpdateStrategyReplace: true,
		UpdateStrategyMerge:   true,
		UpdateStrategyAppend:  true,
	}

	if _, ok := supported[strategy]; !ok {
		return fmt.Errorf("invalid branch_protections_update_strategy: %s (must be 'replace', 'merge', or 'append')", strategy)
	}

	return nil
}
