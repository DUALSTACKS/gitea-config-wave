package cmd

import (
	"fmt"
	"os"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

type BranchProtectionsHandler struct{}

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

	// Clear existing protections
	existing, _, err := client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list existing protections: %w", err)
	}

	for _, bp := range existing {
		if _, err := client.DeleteBranchProtection(owner, repo, bp.BranchName); err != nil {
			return fmt.Errorf("failed to delete protection: %w", err)
		}
	}

	// Create new protections
	for _, bp := range bpConfig.Rules {
		_, _, err := client.CreateBranchProtection(owner, repo, toCreateBranchProtectionOption(bp))
		if err != nil {
			return fmt.Errorf("failed to create protection: %w", err)
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

func (h *BranchProtectionsHandler) Enabled() bool {
	return true
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
