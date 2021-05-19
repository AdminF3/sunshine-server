package controller

import (
	"context"
	"io"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const fmtValidateUser = `{"valid": 3}`

func TestValidateUpdateUser(t *testing.T) {
	e := services.NewTestEnv(t)

	usr := NewUser(e)
	ustore := usr.st.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	u1 := stores.NewTestUser(t, ustore)
	u2 := stores.NewTestUser(t, ustore)
	u3 := stores.NewTestUser(t, ustore)

	cases := []struct {
		name     string
		ctx      context.Context
		user     uuid.UUID
		expected error
		valid    models.ValidationStatus
		body     io.Reader
	}{
		{
			name:     "user validates himself",
			ctx:      services.NewTestContext(t, e, u1),
			user:     u1.ID,
			expected: nil,
			valid:    models.ValidationStatusValid,
			body:     strings.NewReader(fmtValidateUser),
		},
		{
			name:     "admin validates",
			ctx:      services.NewTestContext(t, e, su),
			user:     u3.ID,
			expected: nil,
			valid:    models.ValidationStatusDeclined,
			body:     strings.NewReader(fmtValidateAsset),
		},
		{
			name:     "user validates other user",
			ctx:      services.NewTestContext(t, e, u1),
			user:     u2.ID,
			expected: ErrUnauthorized,
			body:     strings.NewReader(fmtValidateAsset),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			updated, _, err := usr.Update(c.ctx, c.user, c.body)
			if err != c.expected {
				t.Errorf("expected: %v, got: %v", c.expected, err)
			}
			if err == nil && updated.Data.(*models.User).Valid != c.valid {
				t.Errorf("expexted status: %v, got status: %v", c.valid, updated.Data.(*models.User).Valid)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	e := services.NewTestEnv(t)
	mock := gomock.NewController(t)
	defer mock.Finish()

	not := mocks.NewMockNotifier(mock)
	e.Notifier = not
	not.EXPECT().Notify(gomock.Any(), gomock.Any()).AnyTimes()

	u := NewUser(e)
	ustore := u.st.FromKind("user")

	su := stores.NewTestAdmin(t, ustore)
	ca := stores.NewTestPortfolioRole(t, ustore, models.CountryAdminRole, models.CountryLatvia)
	usr := stores.NewTestUser(t, ustore)

	cmnt := "da"
	cases := []struct {
		name      string
		ctx       context.Context
		user      uuid.UUID
		status    models.ValidationStatus
		oldStatus models.ValidationStatus
		expected  error
		comment   *string
	}{
		{
			name:     "ok ca",
			ctx:      services.NewTestContext(t, e, ca),
			user:     usr.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
		},
		{
			name:     "ok su",
			ctx:      services.NewTestContext(t, e, su),
			user:     usr.ID,
			status:   models.ValidationStatusDeclined,
			expected: nil,
		},
		{
			name:      "unauth",
			ctx:       services.NewTestContext(t, e, usr),
			user:      usr.ID,
			status:    models.ValidationStatusValid,
			oldStatus: models.ValidationStatusDeclined,
			expected:  ErrUnauthorized,
		},
		{
			name:     "with comment",
			ctx:      services.NewTestContext(t, e, ca),
			user:     usr.ID,
			status:   models.ValidationStatusRegistered,
			expected: nil,
			comment:  &cmnt,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := u.Validate(c.ctx, c.user, c.status, c.comment)
			if err != c.expected {
				t.Errorf("expected %v err; got %v", c.expected, err)
			}

			upd, _, err := u.Get(c.ctx, c.user)
			if err != nil {
				t.Errorf("could not fetch updated user; got %v", err)
			}
			if c.expected != nil {
				c.status = c.oldStatus
			}
			if upd.Data.(*models.User).Valid != c.status {
				t.Errorf("expected status %v; got %v", c.status, upd.Data.(*models.User).Valid)
			}

		})
	}
}

func TestAdminList(t *testing.T) {

	e := services.NewTestEnv(t)
	u := NewUser(e)
	ustore := u.st.FromKind("user")

	filter := stores.Filter{
		Offset: *new(int),
		Limit:  *new(int),
	}
	// create admin users
	su := stores.NewTestAdmin(t, ustore)
	stores.NewTestAdminNwManager(t, ustore)
	stores.NewTestPlatformManager(t, ustore)
	stores.NewTestPortfolioRole(t, ustore, models.InvestorRole, models.CountryLatvia)
	usr := stores.NewTestUser(t, ustore)

	cases := []struct {
		name     string
		ctx      context.Context
		user     uuid.UUID
		expected error
	}{
		{
			name:     "ok",
			ctx:      services.NewTestContext(t, e, su),
			user:     usr.ID,
			expected: nil,
		},
		{
			name:     "unauth",
			ctx:      services.NewTestContext(t, e, usr),
			user:     usr.ID,
			expected: ErrUnauthorized,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lu, n, err := u.AdminList(c.ctx, filter)
			if err != c.expected {
				t.Errorf("expected %v err; got %v", c.expected, err)
			}
			if err == nil {
				if lu == nil {
					t.Errorf("Got nil users")
				}
				if n != 4 {
					t.Errorf("expected 4 users, got %v", n)
				}
			}
		})
	}

}
