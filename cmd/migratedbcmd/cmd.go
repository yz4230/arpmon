package migratedbcmd

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/fatih/color"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/yz4230/arpmon/internal/db"
)

var MigrateDBCmd = &cobra.Command{
	Use: "migratedb",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := db.LoadConfigFromEnv()
		if err != nil {
			slog.Error("Failed to load database configuration", "error", err)
			return
		}

		attr := color.New(color.FgGreen).SprintFunc()
		cmd.Printf("Target database: host=%s database=%s\n", attr(config.Host), attr(config.Database))

		warning := color.New(color.FgRed, color.Bold).SprintFunc()
		cmd.Print(warning("Do you want to run database migrations? (yes/no): "))
		if !confirmed(cmd.InOrStdin()) {
			slog.Info("Database migration cancelled")
			return
		}

		c, err := pq.NewConnectorConfig(config)
		if err != nil {
			slog.Error("Failed to create database connector", "error", err)
			return
		}

		database := sql.OpenDB(c)
		defer database.Close()

		tx, err := database.BeginTx(cmd.Context(), nil)
		if err != nil {
			slog.Error("Failed to begin database transaction", "error", err)
			return
		}
		defer tx.Rollback()

		if err := db.ApplyMigrations(cmd.Context(), tx); err != nil {
			slog.Error("Failed to apply database migrations", "error", err)
			return
		}

		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit database migrations", "error", err)
			return
		}

		slog.Info("Database migrations completed successfully")
	},
}

func confirmed(r io.Reader) bool {
	var answer string
	_, err := fmt.Fscan(r, &answer)
	return err == nil && strings.EqualFold(answer, "yes")
}
