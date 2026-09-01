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

		return c.Redirect(http.StatusSeeOther, "/members")
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

		return c.Redirect(http.StatusSeeOther, "/members")
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

		return c.Redirect(http.StatusSeeOther, "/members")
	}
	e.POST("/members/:id/delete", deleteMember)
	e.DELETE("/members/:id", deleteMember)
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
		attendances = append(attendances, &domain.MemberAttendance{
			MemberID:  id,
			IsPresent: present[id],
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
