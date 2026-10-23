package gateway

import (
	"fmt"

	"github.com/yz4230/arpmon/internal/domain"
)

type memberAttendanceRow struct {
	MemberName string
	IsPresent  bool
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
