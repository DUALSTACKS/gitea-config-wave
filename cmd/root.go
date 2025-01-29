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
	version = "0.1.0"
	logger  *slog.Logger
)

// Custom handler for colored output
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

// getLevelColor returns the ANSI color code for the given log level
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

	// Format the output: time LEVEL message key=value ...
	fmt.Fprintf(h.out, "%s %s%5s%s %s",
		timeStr,
		color,
		level,
		reset,
		r.Message,
	)

	// Add attributes
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			// Skip time attribute as we already formatted it
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

// rootCmd represents the base command for the gitea-config-wave CLI
var rootCmd = &cobra.Command{
	Use:   "gitea-config-wave",
	Short: "A CLI for synchronizing Gitea repository settings",
	Long:  `A lightweight CLI tool to automate the propagation of repository settings, branch protection rules, and more.`,
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

// Execute runs the root command
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
	// Initialize logger with custom handler for colored output
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

	// Any other global flags would go here
	// ...

	// Subcommands must be added here. For example:
	// rootCmd.AddCommand(pullCmd) // We'll do this from pull.go
}
