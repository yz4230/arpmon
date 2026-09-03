package gateway

import (
	"database/sql"
	"embed"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
	"github.com/yz4230/arpmon/pkg/uptime"
	"github.com/yz4230/arpmon/pkg/usecase"
)

//go:embed static
var staticFiles embed.FS

func RegisterHandlers(e *echo.Echo, i *do.Injector) {
	e.Renderer = &echo.TemplateRenderer{Template: NewTemplateRenderer()}

	e.StaticFS("/static", echo.MustSubFS(staticFiles, "static"))
	registerMemberHandlers(e, i)

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

	newScanRunHandler := func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
		ctx := c.Request().Context()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		members, err := memberRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "new-scan-run", map[string]any{
			"Title":   "New scan run",
			"Members": members,
		})
	}
	e.GET("/scan-runs/new", newScanRunHandler)
	e.GET("/probes/new", newScanRunHandler)

	listScanRunsHandler := func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		scanRunRepository := do.MustInvoke[repository.ScanRunRepository](i)
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
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "scan-runs", map[string]any{
			"Title":    "Scan runs",
			"ScanRuns": scanRuns,
		})
	}
	e.GET("/scan-runs", listScanRunsHandler)
	e.GET("/probes", listScanRunsHandler)

	runScanHandler := func(c *echo.Context) error {
		_, target, err := net.ParseCIDR(strings.TrimSpace(c.FormValue("target")))
		if err != nil {
			return c.String(http.StatusBadRequest, "Target must be a valid CIDR, for example 192.168.1.0/24.")
		}
		members, err := memberAttendancesFromRequest(c.Request())
		if err != nil {
			return c.String(http.StatusBadRequest, "Member selection contains an invalid member ID.")
		}

		scanRunUseCase := do.MustInvoke[usecase.ScanRunUseCase](i)
		scanRun, err := scanRunUseCase.Do(c.Request().Context(), target, members)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to run scan").Wrap(err)
		}
		if scanRun == nil || scanRun.ID <= 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Scan run did not return an ID")
		}
		location := "/scan-runs/" + strconv.Itoa(scanRun.ID)

		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", location)
			return c.NoContent(http.StatusNoContent)
		}

		return c.Redirect(http.StatusSeeOther, location)
	}
	e.POST("/scan-runs", runScanHandler)
	e.POST("/probes", runScanHandler)

	registerScanRunDetailHandler(e, i)

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
