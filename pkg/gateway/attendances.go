package gateway

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

type memberAttendanceRow struct {
	MemberName string
	IsPresent  bool
	CreatedAt  time.Time
}

func registerProbeAttendanceHandler(e *echo.Echo, i *do.Injector) {
	e.GET("/probes/:id/attendances", func(c *echo.Context) error {
		probeID, err := parseProbeID(c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Probe ID must be a positive integer.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		probeRepository := do.MustInvoke[repository.ProbeRepository](i)
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

		return c.Render(http.StatusOK, "attendances", map[string]any{
			"Title":       "Member attendance",
			"Probe":       probe,
			"Attendances": newMemberAttendanceRows(attendances, members, probeID),
		})
	})
}

func parseProbeID(c *echo.Context) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid probe id")
	}
	return id, nil
}

func newMemberAttendanceRows(attendances []*domain.MemberAttendance, members []*domain.Member, probeID int) []memberAttendanceRow {
	memberNames := make(map[int]string, len(members))
	for _, member := range members {
		memberNames[member.ID] = member.Name
	}

	rows := make([]memberAttendanceRow, 0, len(attendances))
	for _, attendance := range attendances {
		if attendance == nil || attendance.ProbeID != probeID {
			continue
		}

		memberName := memberNames[attendance.MemberID]
		if memberName == "" {
			memberName = fmt.Sprintf("Member #%d", attendance.MemberID)
		}
		rows = append(rows, memberAttendanceRow{
			MemberName: memberName,
			IsPresent:  attendance.IsPresent,
			CreatedAt:  attendance.CreatedAt,
		})
	}

	return rows
}
