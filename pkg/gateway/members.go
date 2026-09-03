package gateway

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

func registerMemberHandlers(e *echo.Echo, i *do.Injector) {
	e.GET("/members", func(c *echo.Context) error {
		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)

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

		return c.Render(http.StatusOK, "members", map[string]any{
			"Title":   "Members",
			"Members": members,
			"Message": memberSuccessMessage(c.QueryParam("status")),
		})
	})

	e.POST("/members", func(c *echo.Context) error {
		name := strings.TrimSpace(c.FormValue("name"))
		if name == "" {
			return c.String(http.StatusBadRequest, "Member name is required.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err := memberRepository.Create(ctx, tx, &domain.Member{Name: name}); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/members?status=created")
	})

	updateMember := func(c *echo.Context) error {
		id, err := parseMemberID(c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Member ID must be a positive integer.")
		}

		name := strings.TrimSpace(c.FormValue("name"))
		if name == "" {
			return c.String(http.StatusBadRequest, "Member name is required.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err := memberRepository.Update(ctx, tx, &domain.Member{ID: id, Name: name}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusNotFound, "Member not found")
			}
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/members?status=updated")
	}
	e.POST("/members/:id", updateMember)
	e.PUT("/members/:id", updateMember)

	deleteMember := func(c *echo.Context) error {
		id, err := parseMemberID(c)
		if err != nil {
			return c.String(http.StatusBadRequest, "Member ID must be a positive integer.")
		}

		ctx := c.Request().Context()
		db := do.MustInvoke[*sql.DB](i)
		memberRepository := do.MustInvoke[repository.MemberRepository](i)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := memberRepository.DeleteByID(ctx, tx, id); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/members?status=deleted")
	}
	e.POST("/members/:id/delete", deleteMember)
	e.DELETE("/members/:id", deleteMember)
}

func memberSuccessMessage(status string) string {
	switch status {
	case "created":
		return "Member created successfully. It will appear in the attendance checklist before the next scan."
	case "updated":
		return "Member updated successfully. Future attendance checklists will use the new name."
	case "deleted":
		return "Member deleted successfully. It has been removed from future attendance checklists."
	default:
		return ""
	}
}

func parseMemberID(c *echo.Context) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(c.Param("id")))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid member id")
	}
	return id, nil
}

func memberAttendancesFromRequest(r *http.Request) ([]*domain.MemberAttendance, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	memberIDs, err := parseMemberIDs(r.Form["member_ids"])
	if err != nil {
		return nil, err
	}
	presentMemberIDs, err := parseMemberIDs(r.Form["present_member_ids"])
	if err != nil {
		return nil, err
	}
	presenceByMemberID, err := parseMemberPresence(r.Form)
	if err != nil {
		return nil, err
	}
	if len(memberIDs) == 0 {
		memberIDs = presentMemberIDs
	}

	present := make(map[int]bool, len(presentMemberIDs))
	for _, id := range presentMemberIDs {
		present[id] = true
	}

	seen := make(map[int]struct{}, len(memberIDs))
	attendances := make([]*domain.MemberAttendance, 0, len(memberIDs))
	for _, id := range memberIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		isPresent := present[id]
		if selected, ok := presenceByMemberID[id]; ok {
			isPresent = selected
		}
		attendances = append(attendances, &domain.MemberAttendance{
			MemberID:  id,
			IsPresent: isPresent,
		})
	}

	return attendances, nil
}

func parseMemberIDs(values []string) ([]int, error) {
	ids := make([]int, 0, len(values))
	for _, value := range values {
		id, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid member id %q", value)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func parseMemberPresence(values map[string][]string) (map[int]bool, error) {
	const prefix = "member_presence_"

	presenceByMemberID := make(map[int]bool)
	for key, rawValues := range values {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		id, err := strconv.Atoi(strings.TrimPrefix(key, prefix))
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid member id in presence selection %q", key)
		}
		if len(rawValues) != 1 {
			return nil, fmt.Errorf("member presence selection for %d must contain one value", id)
		}

		switch strings.ToLower(strings.TrimSpace(rawValues[0])) {
		case "present", "true", "1":
			presenceByMemberID[id] = true
		case "not present", "not_present", "false", "0":
			presenceByMemberID[id] = false
		default:
			return nil, fmt.Errorf("invalid presence selection for member %d", id)
		}
	}

	return presenceByMemberID, nil
}
