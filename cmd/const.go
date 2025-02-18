package cmd

const (
	DefaultConfigFile                   = "gitea-config-wave.yaml"
	DefaultOutputDir                    = ".gitea/defaults"
	DefaultRepoSettingsFile             = "repo_settings.yaml"
	DefaultBranchProtectionsFile        = "branch_protections.yaml"
	DefaultTagProtectionsFile           = "tag_protections.yaml"
	DefaultWebhooksFile                 = "webhooks.yaml"
	DefaultTopicsFile                   = "topics.yaml"
	DefaultTemplatesFile                = "templates.yaml"
	DefaultTemplatesUpdateBranchName    = "gitea-config-wave/sync-templates"
	DefaultTemplatesUpdateCommitMessage = "chore(docs): update PR and issue templates"
	DefaultTemplatesUpdatePRDescription = `# Gitea Config Wave - Issue/PR Templates Sync
This PR is automatically created by [Gitea Config Wave](https://github.com/dualstacks/gitea-config-wave) to sync issue and PR templates.
`
	DefaultTopicsUpdateStrategy            = UpdateStrategyAppend
	DefaultBranchProtectionsUpdateStrategy = UpdateStrategyAppend
	DefaultTagProtectionsUpdateStrategy    = UpdateStrategyAppend
	DefaultWebhooksUpdateStrategy          = UpdateStrategyAppend
	DefaultTemplatesUpdateStrategy         = UpdateStrategyReplace
)

type UpdateStrategy string

const (
	UpdateStrategyReplace UpdateStrategy = "replace"
	UpdateStrategyMerge   UpdateStrategy = "merge"
	UpdateStrategyAppend  UpdateStrategy = "append"
)
