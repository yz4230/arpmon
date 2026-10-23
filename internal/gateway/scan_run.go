package gateway

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/repository"
	"github.com/yz4230/arpmon/internal/usecase"
)

const defaultScanRunTargetNetwork = "192.168.1.0/24"

func registerScanRunHandlers(e *echo.Echo, i *do.Injector) {
	e.GET("/scan-runs/new", func(c *echo.Context) error {
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
		scanRunRepository := do.MustInvoke[repository.ScanRunRepository](i)
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
		scanRuns, err := scanRunRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		targetNetwork := defaultScanRunTargetNetwork
		if latest := latestScanRun(scanRuns); latest != nil && latest.TargetNetwork != nil {
			targetNetwork = latest.TargetNetwork.String()
		}

		return c.Render(http.StatusOK, "new-scan-run", map[string]any{
			"Title":         "New scan run",
			"Members":       members,
			"TargetNetwork": targetNetwork,
		})
	})

	e.GET("/scan-runs", func(c *echo.Context) error {
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
	})

	e.POST("/scan-runs", func(c *echo.Context) error {
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
	})

	e.GET("/scan-runs/:id", func(c *echo.Context) error {
		scanRunID, err := parseScanRunID(c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Scan run ID must be a positive integer.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		scanRunRepository := do.MustInvoke[repository.ScanRunRepository](i)
		deviceRepository := do.MustInvoke[repository.DeviceRepository](i)
		deviceObservationRepository := do.MustInvoke[repository.DeviceObservationRepository](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
		memberAttendanceRepository := do.MustInvoke[repository.MemberAttendanceRepository](i)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		scanRun, err := scanRunRepository.GetByID(ctx, tx, scanRunID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "Scan run not found")
			}
			return err
		}
		deviceObservations, err := deviceObservationRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		devices, err := deviceRepository.List(ctx, tx)
		if err != nil {
			return err
		}
		attendances, err := memberAttendanceRepository.List(ctx, tx)
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

		return c.Render(http.StatusOK, "scan-run", map[string]any{
			"Title":              "Scan run",
			"ScanRun":            scanRun,
			"DeviceObservations": newDeviceObservationRows(deviceObservations, devices, scanRunID),
			"Attendances":        newMemberAttendanceRows(attendances, members, scanRunID),
		})
	})
}

func parseScanRunID(c *echo.Context) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid scan run id")
	}
	return id, nil
}

func latestScanRun(scanRuns []*domain.ScanRun) *domain.ScanRun {
	var latest *domain.ScanRun
	for _, scanRun := range scanRuns {
		if scanRun == nil {
			continue
		}
		if latest == nil || scanRun.CreatedAt.After(latest.CreatedAt) ||
			(scanRun.CreatedAt.Equal(latest.CreatedAt) && scanRun.ID > latest.ID) {
			latest = scanRun
		}
	}
	return latest
}
