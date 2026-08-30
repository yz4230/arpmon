package gateway

import (
	"database/sql"
	"embed"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/repository"
	"github.com/yz4230/arpmon/pkg/uptime"
	"github.com/yz4230/arpmon/pkg/usecase"
)

//go:embed static
var staticFiles embed.FS

func RegisterHandlers(e *echo.Echo, i *do.Injector) {
	e.Renderer = &echo.TemplateRenderer{Template: NewTemplateRenderer()}

	e.StaticFS("/static", echo.MustSubFS(staticFiles, "static"))

	e.GET("/health", func(c *echo.Context) error {
		var health struct {
			Uptime string `json:"uptime"`
			DB     string `json:"db"`
		}

		health.Uptime = uptime.Get().String()

		conn := do.MustInvoke[*sql.DB](i)
		if err := conn.Ping(); err != nil {
			health.DB = "unhealthy"
		} else {
			health.DB = "healthy"
		}

		return c.JSON(http.StatusOK, health)
	})

	e.POST("/probes", func(c *echo.Context) error {
		_, target, err := net.ParseCIDR(strings.TrimSpace(c.FormValue("target")))
		if err != nil {
			return c.String(http.StatusBadRequest, "Target must be a valid CIDR, for example 192.168.1.0/24.")
		}

		probeUseCase := do.MustInvoke[usecase.ProbeUseCase](i)
		if err := probeUseCase.Do(c.Request().Context(), target); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to run probe").Wrap(err)
		}

		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Refresh", "true")
			return c.NoContent(http.StatusNoContent)
		}

		return c.Redirect(http.StatusSeeOther, "/")
	})

	e.GET("/", func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		probeRepository := do.MustInvoke[repository.ProbeRepository](i)
		ctx := c.Request().Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		probes, err := probeRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "index", map[string]any{
			"Title":  "ARP Monitor",
			"Probes": probes,
		})
	})
}
