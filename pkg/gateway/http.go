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
	registerProbeAttendanceHandler(e, i)

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

	e.GET("/probes/new", func(c *echo.Context) error {
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

		return c.Render(http.StatusOK, "new-probe", map[string]any{
			"Title":   "New probe",
			"Members": members,
		})
	})

	e.GET("/probes", func(c *echo.Context) error {
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

		return c.Render(http.StatusOK, "probes", map[string]any{
			"Title":  "Probes",
			"Probes": probes,
		})
	})

	e.POST("/probes", func(c *echo.Context) error {
		_, target, err := net.ParseCIDR(strings.TrimSpace(c.FormValue("target")))
		if err != nil {
			return c.String(http.StatusBadRequest, "Target must be a valid CIDR, for example 192.168.1.0/24.")
		}
		members, err := memberAttendancesFromRequest(c.Request())
		if err != nil {
			return c.String(http.StatusBadRequest, "Member selection contains an invalid member ID.")
		}

		probeUseCase := do.MustInvoke[usecase.ProbeUseCase](i)
		probe, err := probeUseCase.Do(c.Request().Context(), target, members)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to run probe").Wrap(err)
		}
		if probe == nil || probe.ID <= 0 {
			return echo.NewHTTPError(http.StatusInternalServerError, "Probe did not return an ID")
		}
		location := "/probes/" + strconv.Itoa(probe.ID)

		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Redirect", location)
			return c.NoContent(http.StatusNoContent)
		}

		return c.Redirect(http.StatusSeeOther, location)
	})

	registerProbeDetailHandler(e, i)

	e.GET("/", func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		probeRepository := do.MustInvoke[repository.ProbeRepository](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
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
		members, err := memberRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Render(http.StatusOK, "index", map[string]any{
			"Title": "ARP Monitor",
			"Stats": newDashboardStats(members, probes, formatUptime(uptime.Get())),
		})
	})
}

type DashboardStats struct {
	Uptime          string
	TotalProbeCount int
	MemberCount     int
}

func formatUptime(value time.Duration) string {
	return value.Round(time.Second).String()
}

func newDashboardStats(members []*domain.Member, probes []*domain.Probe, uptimeText string) DashboardStats {
	return DashboardStats{
		Uptime:          uptimeText,
		TotalProbeCount: len(probes),
		MemberCount:     len(members),
	}
}
