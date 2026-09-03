package gateway

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/yz4230/arpmon/pkg/domain"
)

type memberAttendanceRow struct {
	MemberName string
	IsPresent  bool
}

func parseScanRunID(c *echo.Context) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid scan run id")
	}
	return id, nil
}

func newMemberAttendanceRows(attendances []*domain.MemberAttendance, members []*domain.Member, scanRunID int) []memberAttendanceRow {
	memberNames := make(map[int]string, len(members))
	for _, member := range members {
		memberNames[member.ID] = member.Name
	}

	rows := make([]memberAttendanceRow, 0, len(attendances))
	for _, attendance := range attendances {
		if attendance == nil || attendance.ScanRunID != scanRunID {
			continue
		}

		memberName := memberNames[attendance.MemberID]
		if memberName == "" {
			memberName = fmt.Sprintf("Member #%d", attendance.MemberID)
		}
		rows = append(rows, memberAttendanceRow{
			MemberName: memberName,
			IsPresent:  attendance.IsPresent,
		})
	}

	return rows
}
