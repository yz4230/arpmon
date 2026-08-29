package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/yz4230/arpmon/cmd/migratedbcmd"
	"github.com/yz4230/arpmon/cmd/pingdbcmd"
	"github.com/yz4230/arpmon/cmd/scancmd"
	"github.com/yz4230/arpmon/cmd/servecmd"
)

var rootPstFlags struct {
	verbose bool
}

var rootCmd = &cobra.Command{
	Use: filepath.Base(os.Args[0]),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		options := &tint.Options{AddSource: true, TimeFormat: time.TimeOnly}
		if rootPstFlags.verbose {
			options.Level = slog.LevelDebug
		}
		logger := slog.New(tint.NewTextHandler(os.Stderr, options))
		slog.SetDefault(logger)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(migratedbcmd.MigrateDBCmd)
	rootCmd.AddCommand(pingdbcmd.PingDBCmd)
	rootCmd.AddCommand(scancmd.ScanCmd)
	rootCmd.AddCommand(servecmd.ServeCmd)

	rootCmd.PersistentFlags().BoolVarP(&rootPstFlags.verbose, "verbose", "v", false, "Enable verbose output")
}
