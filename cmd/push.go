package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
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

		cfg, err := LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client, err := gitea.NewClient(cfg.GiteaURL, gitea.SetToken(cfg.GiteaToken))
		if err != nil {
			return fmt.Errorf("failed to create Gitea client: %w", err)
		}

		targetRepos, err := getAllTargetRepos(cmd, client, cfg, args)
		if err != nil {
			return err
		}
		if len(targetRepos) == 0 {
			return errors.New("no repositories to process after merges/exclusions")
		}

		logger.Info("found target repositories", "count", len(targetRepos))
		for _, repo := range targetRepos {
			logger.Info("📦 " + repo)
		}

		outputDir := cfg.Config.OutputDir
		if outputDir == "" {
			outputDir = DefaultOutputDir
		}

		// Initialize handlers based on push configuration
		var handlers []ConfigHandler
		if cfg.Push.RepoSettings {
			handlers = append(handlers, &RepoSettingsHandler{})
		}
		if cfg.Push.Topics {
			handlers = append(handlers, &TopicsHandler{})
		}
		if cfg.Push.BranchProtections {
			handlers = append(handlers, &BranchProtectionsHandler{})
		}
		if cfg.Push.Webhooks {
			handlers = append(handlers, &WebhooksHandler{})
		}

		if len(handlers) == 0 {
			logger.Info("🤷 no items enabled in push config - nothing to do")
			return nil
		}

		for _, fullName := range targetRepos {
			owner, repo, err := parseRepoString(fullName)
			if err != nil {
				return fmt.Errorf("invalid repo argument %q: %w", fullName, err)
			}

			if dryRun || cfg.DryRun {
				logger.Info("(dry run) will apply settings to",
					"owner", owner,
					"repo", repo,
				)
				continue
			}

			// Process each handler for this repository
			for _, handler := range handlers {
				if !handler.Enabled() {
					continue
				}

				logger.Debug("processing handler",
					"handler", handler.Name(),
					"owner", owner,
					"repo", repo,
				)

				// Load handler data from file
				data, err := handler.Load(filepath.Join(outputDir, handler.Path()))
				if err != nil {
					return fmt.Errorf("failed to load data for handler %s: %w", handler.Name(), err)
				}

				// Push changes using handler
				err = handler.Push(client, owner, repo, data)
				if err != nil {
					return fmt.Errorf("failed to push %s for %s/%s: %w", handler.Name(), owner, repo, err)
				}

				logger.Debug("successfully processed handler",
					"handler", handler.Name(),
					"owner", owner,
					"repo", repo,
				)
			}

			logger.Info("successfully pushed settings",
				"owner", owner,
				"repo", repo,
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
