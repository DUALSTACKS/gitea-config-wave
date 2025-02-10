package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// pullCmd handles pulling repository settings from Gitea instances
var pullCmd = &cobra.Command{
	Use:   "pull [owner/repo]",
	Short: "Pull settings from a Gitea repo",
	Long: `Pulls repository settings (e.g., branch protections,
issues/PR templates, etc.) from a specified Gitea repository and 
saves them to YAML files in the output directory (defaults to .gitea/defaults).`,
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

		// Initialize handlers based on pull configuration
		var handlers []ConfigHandler
		if cfg.Pull.RepoSettings {
			handlers = append(handlers, &RepoSettingsHandler{})
		}
		if cfg.Pull.Topics {
			handlers = append(handlers, &TopicsHandler{})
		}
		if cfg.Pull.BranchProtections {
			handlers = append(handlers, &BranchProtectionsHandler{})
		}
		if cfg.Pull.Webhooks {
			handlers = append(handlers, &WebhooksHandler{})
		}
		// if cfg.Pull.TagProtections {
		// 	handlers = append(handlers, &TagProtectionsHandler{})
		// }

		if len(handlers) == 0 {
			logger.Info("🤷 no items enabled in pull config - nothing to do")
			return nil
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
			)
			return nil
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
		}

		for _, handler := range handlers {
			if !handler.Enabled() {
				continue
			}

			logger.Debug("pulling configuration",
				"handler", handler.Name(),
				"owner", owner,
				"repo", repo,
			)

			data, err := handler.Pull(client, owner, repo)
			if err != nil {
				return fmt.Errorf("failed to pull %s: %w", handler.Name(), err)
			}

			outputPath := filepath.Join(outputDir, handler.Path())
			if err := WriteYAMLFile(outputPath, data); err != nil {
				return fmt.Errorf("failed to write %s: %w", handler.Name(), err)
			}
		}

		logger.Info("successfully pulled repository settings",
			"owner", owner,
			"repo", repo,
			"output_dir", outputDir,
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
