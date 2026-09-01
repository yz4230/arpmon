package gateway

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
	"github.com/yz4230/arpmon/pkg/testutil"
	"github.com/yz4230/arpmon/pkg/usecase"
)

type memberRepositoryStub struct {
	nextID  int
	members []*domain.Member
}

func (r *memberRepositoryStub) Create(_ context.Context, _ *sql.Tx, member *domain.Member) (*domain.Member, error) {
	created := *member
	if r.nextID == 0 {
		r.nextID = 1
	}
	created.ID = r.nextID
	r.nextID++
	r.members = append(r.members, &created)
	return &created, nil
}

func (r *memberRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.Member, error) {
	for _, member := range r.members {
		if member.ID == id {
			copy := *member
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memberRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.Member, error) {
	members := make([]*domain.Member, 0, len(r.members))
	for _, member := range r.members {
		copy := *member
		members = append(members, &copy)
	}
	return members, nil
}

func (r *memberRepositoryStub) Update(_ context.Context, _ *sql.Tx, member *domain.Member) (*domain.Member, error) {
	for i, current := range r.members {
		if current.ID == member.ID {
			updated := *member
			r.members[i] = &updated
			return &updated, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *memberRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, member := range r.members {
		if member.ID == id {
			r.members = append(r.members[:i], r.members[i+1:]...)
			return nil
		}
	}
	return nil
}

var _ repository.MemberRepository = (*memberRepositoryStub)(nil)

func newGatewayWithMemberRepository(t *testing.T, memberRepository repository.MemberRepository) (*echo.Echo, *testutil.InMemoryDB) {
	t.Helper()

	tdb := testutil.NewInMemoryDB()
	injector := do.New()
	do.ProvideValue(injector, tdb.Conn)
	do.ProvideValue(injector, memberRepository)
	do.ProvideValue(injector, usecase.ProbeUseCase(&probeUseCaseStub{}))
	e := echo.New()
	RegisterHandlers(e, injector)
	return e, tdb
}

func formRequest(method, target string, values url.Values) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestMemberCRUDRoutes(t *testing.T) {
	memberRepository := &memberRepositoryStub{}
	e, tdb := newGatewayWithMemberRepository(t, memberRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	createRecorder := httptest.NewRecorder()
	e.ServeHTTP(createRecorder, formRequest(http.MethodPost, "/members", url.Values{"name": {" Alice "}}))
	if createRecorder.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", createRecorder.Code, http.StatusSeeOther)
	}
	if len(memberRepository.members) != 1 || memberRepository.members[0].Name != "Alice" {
		t.Fatalf("members after create = %#v, want Alice", memberRepository.members)
	}

	listRecorder := httptest.NewRecorder()
	e.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/members", nil))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRecorder.Code, http.StatusOK)
	}
	if !strings.Contains(listRecorder.Body.String(), "Alice") {
		t.Fatalf("member list body = %q, want Alice", listRecorder.Body.String())
	}

	updateRecorder := httptest.NewRecorder()
	e.ServeHTTP(updateRecorder, formRequest(http.MethodPost, "/members/1", url.Values{"name": {"Bob"}}))
	if updateRecorder.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d", updateRecorder.Code, http.StatusSeeOther)
	}
	if memberRepository.members[0].Name != "Bob" {
		t.Fatalf("member after update = %#v, want Bob", memberRepository.members[0])
	}

	deleteRecorder := httptest.NewRecorder()
	e.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodPost, "/members/1/delete", nil))
	if deleteRecorder.Code != http.StatusSeeOther {
		t.Fatalf("delete status = %d, want %d", deleteRecorder.Code, http.StatusSeeOther)
	}
	if len(memberRepository.members) != 0 {
		t.Fatalf("members after delete = %#v, want empty", memberRepository.members)
	}
}

func TestMemberAttendancesFromRequest(t *testing.T) {
	form := url.Values{
		"member_ids":         {"1", "2", "2"},
		"present_member_ids": {"2"},
	}
	r := formRequest(http.MethodPost, "/probes", form)

	got, err := memberAttendancesFromRequest(r)
	if err != nil {
		t.Fatalf("memberAttendancesFromRequest() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("memberAttendancesFromRequest() returned %d attendances, want 2", len(got))
	}
	if got[0].MemberID != 1 || got[0].IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got[0])
	}
	if got[1].MemberID != 2 || !got[1].IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got[1])
	}
}

func TestMemberAttendancesFromRequestParsesPresenceChoices(t *testing.T) {
	form := url.Values{
		"member_ids":        {"1", "2"},
		"member_presence_1": {"Not Present"},
		"member_presence_2": {"Present"},
	}
	r := formRequest(http.MethodPost, "/probes", form)

	got, err := memberAttendancesFromRequest(r)
	if err != nil {
		t.Fatalf("memberAttendancesFromRequest() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("memberAttendancesFromRequest() returned %d attendances, want 2", len(got))
	}
	if got[0].MemberID != 1 || got[0].IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got[0])
	}
	if got[1].MemberID != 2 || !got[1].IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got[1])
	}
}
