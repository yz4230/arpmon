package pingdbcmd

import (
	"database/sql"
	"log/slog"

	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/yz4230/arpmon/internal/db"
)

var PingDBCmd = &cobra.Command{
	Use: "pingdb",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := db.LoadConfigFromEnv()
		if err != nil {
			slog.Error("Failed to load database configuration", "error", err)
			return
		}

		c, err := pq.NewConnectorConfig(config)
		if err != nil {
			slog.Error("Failed to create database connector", "error", err)
			return
		}

		db := sql.OpenDB(c)
		defer db.Close()

		err = db.Ping()
		if err != nil {
			slog.Error("Failed to connect to the database", "error", err)
			return
		}

		slog.Info("Successfully connected to the database")
	},
}
