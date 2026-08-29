package servecmd

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lib/pq"
	"github.com/samber/do"
	"github.com/spf13/cobra"
	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/db"
	"github.com/yz4230/arpmon/pkg/gateway"
	"github.com/yz4230/arpmon/pkg/repository"
)

var ServeCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		port := serverPort()

		slog.Info("Starting server", "port", port)

		e := echo.New()
		e.Logger = slog.Default()

		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:   true,
			LogURI:      true,
			LogLatency:  true,
			LogMethod:   true,
			HandleError: true,
			LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
				args := []any{
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
				}
				if v.Error == nil {
					slog.Info("REQUEST", args...)
				} else {
					args = append(args, "error", v.Error)
					slog.Error("REQUEST_ERROR", args...)
				}
				return nil
			},
		}))

		e.GET("/ping", func(c *echo.Context) error {
			return c.String(http.StatusOK, "pong")
		})

		injector := do.New()
		injectDependencies(injector)
		gateway.RegisterHandlers(e, injector)

		if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	},
}

func serverPort() int {
	text := os.Getenv("PORT")
	if port, err := strconv.Atoi(text); err == nil {
		return port
	}
	return 8080 // default port
}

func connectDatabase() (*sql.DB, error) {
	config, err := db.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	c, err := pq.NewConnectorConfig(config)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(c)
	return db, nil
}

func injectDependencies(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (*sql.DB, error) {
		db, err := connectDatabase()
		if err != nil {
			slog.Error("Failed to connect to the database", "error", err)
		}
		return db, err
	})

	do.ProvideValue(i, arpscan.NewARPScanner())
	do.ProvideValue(i, repository.NewMemberRepository())
	do.ProvideValue(i, repository.NewMemberAttendanceRepository())
	do.ProvideValue(i, repository.NewProbeRepository())
	do.ProvideValue(i, repository.NewDeviceRepository())
	do.ProvideValue(i, repository.NewDeviceProbeRepository())
}
