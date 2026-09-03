package gateway

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/repository"
)

func registerScanRunDetailHandler(e *echo.Echo, i *do.Injector) {
	handler := func(c *echo.Context) error {
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
	}
	e.GET("/scan-runs/:id", handler)
	e.GET("/probes/:id", handler)
}
