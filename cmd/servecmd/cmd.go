package servecmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/cobra"
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
