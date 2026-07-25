package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	debug     bool
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:   "c3vm",
	Short: "C3 language version manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return setupLogger()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text|json)")
}

func setupLogger() error {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler

	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case "text":
		handler = slog.NewTextHandler(newColorWriter(os.Stderr), opts)
	default:
		return fmt.Errorf("unsupported log format: %s (use text or json)", logFormat)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}
