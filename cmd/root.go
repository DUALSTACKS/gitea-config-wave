package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	logger  *slog.Logger
)

var Version = "dev"

type colorHandler struct {
	out    io.Writer
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

func newColorHandler(w io.Writer, opts *slog.HandlerOptions) *colorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &colorHandler{
		out:  w,
		opts: opts,
	}
}

func getLevelColor(level slog.Level) string {
	switch level {
	case slog.LevelError:
		return "\033[31m" // Red
	case slog.LevelWarn:
		return "\033[33m" // Yellow
	case slog.LevelInfo:
		return "\033[32m" // Green
	case slog.LevelDebug:
		return "\033[36m" // Cyan
	default:
		return "\033[0m" // Reset
	}
}

func (h *colorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := h.opts.Level
	if minLevel == nil {
		minLevel = slog.LevelInfo
	}
	return level >= minLevel.Level()
}

func (h *colorHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	timeStr := r.Time.Format("15:04:05")
	color := getLevelColor(r.Level)
	reset := "\033[0m"

	fmt.Fprintf(h.out, "%s %s%5s%s %s",
		timeStr,
		color,
		level,
		reset,
		r.Message,
	)

	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == slog.TimeKey {
				return true
			}
			fmt.Fprintf(h.out, " %s=%v", a.Key, a.Value)
			return true
		})
	}
	fmt.Fprintln(h.out)
	return nil
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &colorHandler{
		out:    h.out,
		opts:   h.opts,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return &colorHandler{
		out:    h.out,
		opts:   h.opts,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

var rootCmd = &cobra.Command{
	Use:     "gitea-config-wave",
	Short:   "A CLI for synchronizing Gitea repository settings",
	Long:    `A lightweight CLI tool to automate the propagation of repository settings, branch protection rules, and more.`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			opts := &slog.HandlerOptions{Level: slog.LevelDebug}
			handler := newColorHandler(os.Stdout, opts)
			newLogger := slog.New(handler)
			slog.SetDefault(newLogger)
			logger = newLogger
			logger.Debug("verbose logging enabled")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("failed to execute command",
			"error", err,
			"cmd", os.Args[0],
		)
		os.Exit(1)
	}
}

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := newColorHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)

	// Global --config flag
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"Path to the config file (e.g. gitea-config-wave.yaml)")

	// Global --dry-run flag
	rootCmd.PersistentFlags().Bool("dry-run", false,
		"Show what would happen without making changes")

	// Add --verbose flag for debug logging
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
}

type Config struct {
	GiteaURL   string `yaml:"gitea_url" validate:"required,url"`
	GiteaToken string `yaml:"gitea_token" validate:"required"`
	Config     struct {
		OutputDir string `yaml:"output_dir" validate:"omitempty,dirpath"`
	} `yaml:"config"`
	Pull struct {
		RepoSettings      bool `yaml:"repo_settings"`
		Topics            bool `yaml:"topics"`
		BranchProtections bool `yaml:"branch_protections"`
		TagProtections    bool `yaml:"tag_protections"`
		Webhooks          bool `yaml:"webhooks"`
		Templates         bool `yaml:"templates"`
	} `yaml:"pull"`
	Push struct {
		RepoSettings      bool `yaml:"repo_settings"`
		Topics            bool `yaml:"topics"`
		BranchProtections bool `yaml:"branch_protections"`
		TagProtections    bool `yaml:"tag_protections"`
		Webhooks          bool `yaml:"webhooks"`
		Templates         bool `yaml:"templates"`
	} `yaml:"push"`
	Targets struct {
		Autodiscover       bool     `yaml:"autodiscover"`
		Organization       string   `yaml:"organization"`
		AutodiscoverFilter string   `yaml:"autodiscover_filter"`
		Repos              []string `yaml:"repos"`
		ExcludeRepos       []string `yaml:"exclude_repos"`
	} `yaml:"targets"`
	DryRun                          bool           `yaml:"dry_run"`
	TopicsUpdateStrategy            UpdateStrategy `yaml:"topics_update_strategy"`
	BranchProtectionsUpdateStrategy UpdateStrategy `yaml:"branch_protections_update_strategy"`
	TagProtectionsUpdateStrategy    UpdateStrategy `yaml:"tag_protections_update_strategy"`
	WebhooksUpdateStrategy          UpdateStrategy `yaml:"webhooks_update_strategy"`
}
