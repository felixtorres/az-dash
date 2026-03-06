package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/felixtorres/az-dash/internal/config"
	"github.com/felixtorres/az-dash/internal/tui"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "az-dash",
	Short: "A terminal dashboard for Azure DevOps",
	Long:  "A rich terminal UI for Azure DevOps — pull requests, work items, and pipelines at a glance.",
	RunE:  run,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.az-dash.yml)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfgPath := cfgFile
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		cfgPath = filepath.Join(home, ".az-dash.yml")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	return tui.Start(cfg)
}
