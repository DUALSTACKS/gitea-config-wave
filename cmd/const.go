package cmd

const (
	DefaultConfigFile                      = "gitea-config-wave.yaml"
	DefaultOutputDir                       = ".gitea/defaults"
	DefaultRepoSettingsFile                = "repo_settings.yaml"
	DefaultBranchProtectionsFile           = "branch_protections.yaml"
	DefaultTagProtectionsFile              = "tag_protections.yaml"
	DefaultWebhooksFile                    = "webhooks.yaml"
	DefaultTopicsFile                      = "topics.yaml"
	DefaultTopicsUpdateStrategy            = UpdateStrategyAppend
	DefaultBranchProtectionsUpdateStrategy = UpdateStrategyAppend
	DefaultTagProtectionsUpdateStrategy    = UpdateStrategyAppend
	DefaultWebhooksUpdateStrategy          = UpdateStrategyAppend
)

type UpdateStrategy string

const (
	UpdateStrategyReplace UpdateStrategy = "replace"
	UpdateStrategyMerge   UpdateStrategy = "merge"
	UpdateStrategyAppend  UpdateStrategy = "append"
)
