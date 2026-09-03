package gateway

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/repository"
)

func registerProbeDetailHandler(e *echo.Echo, i *do.Injector) {
	e.GET("/probes/:id", func(c *echo.Context) error {
		probeID, err := parseProbeID(c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Probe ID must be a positive integer.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		probeRepository := do.MustInvoke[repository.ProbeRepository](i)
		deviceRepository := do.MustInvoke[repository.DeviceRepository](i)
		deviceProbeRepository := do.MustInvoke[repository.DeviceProbeRepository](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)
		memberAttendanceRepository := do.MustInvoke[repository.MemberAttendanceRepository](i)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		probe, err := probeRepository.GetByID(ctx, tx, probeID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "Probe not found")
			}
			return err
		}
		deviceProbes, err := deviceProbeRepository.List(ctx, tx)
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

		return c.Render(http.StatusOK, "probe", map[string]any{
			"Title":        "Probe",
			"Probe":        probe,
			"DeviceProbes": newDeviceProbeRows(deviceProbes, devices, probeID),
			"Attendances":  newMemberAttendanceRows(attendances, members, probeID),
		})
	})
}
