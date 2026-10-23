package gateway

import (
	"database/sql"
	"embed"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/repository"
	"github.com/yz4230/arpmon/internal/uptime"
)

//go:embed static
var staticFiles embed.FS

func RegisterHandlers(e *echo.Echo, i *do.Injector) {
	e.Renderer = &echo.TemplateRenderer{Template: NewTemplateRenderer()}

	e.StaticFS("/static", echo.MustSubFS(staticFiles, "static"))
	registerMemberHandlers(e, i)
	registerScanRunHandlers(e, i)

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

	e.GET("/", func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		scanRunRepository := do.MustInvoke[repository.ScanRunRepository](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
		ctx := c.Request().Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		scanRuns, err := scanRunRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		members, err := memberRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "index", map[string]any{
			"Title": "ARP Monitor",
			"Stats": newDashboardStats(members, scanRuns, formatUptime(uptime.Get())),
		})
	})
}

type DashboardStats struct {
	Uptime            string
	TotalScanRunCount int
	MemberCount       int
}

func formatUptime(value time.Duration) string {
	return value.Round(time.Second).String()
}

func newDashboardStats(members []*domain.Member, scanRuns []*domain.ScanRun, uptimeText string) DashboardStats {
	return DashboardStats{
		Uptime:            uptimeText,
		TotalScanRunCount: len(scanRuns),
		MemberCount:       len(members),
	}
}
