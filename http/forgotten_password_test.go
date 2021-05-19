package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/mocks"
	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var (
	any = gomock.Any()
	ctx = context.Background()
)

func TestForgottenPassword(t *testing.T) {
	t.Run("declare", func(t *testing.T) { testForgottenPasswordDeclare(t) })
	t.Run("confirm", func(t *testing.T) { testForgottenPasswordConfirm(t) })
	t.Run("change", func(t *testing.T) { testForgottenPasswordChange(t) })
}

func testForgottenPasswordDeclare(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	mock := gomock.NewController(t)
	defer mock.Finish()

	userdoc := stores.NewTestUser(t, e.UserStore)
	user := userdoc.Data.(*models.User)
	tt := []struct {
		body   string
		status int
	}{
		{body: user.Email, status: http.StatusOK},
		{body: `"` + user.Email + `"`, status: http.StatusOK},
		{body: user.Name, status: http.StatusBadRequest},
		{body: "", status: http.StatusBadRequest},
		{body: "notexisting@example.com", status: http.StatusBadRequest},
	}

	for _, tc := range tt {
		t.Run(tc.body, func(t *testing.T) {
			ts := mocks.NewMockTokenStore(mock)
			m := mocks.NewMockMailer(mock)
			e.TokenStore = ts
			e.Mailer = m
			router := New(e)

			r := httptest.NewRequest("POST",
				"/auth/forgotten_password",
				strings.NewReader(tc.body),
			)

			if tc.status == http.StatusOK {
				token := models.NewToken(models.ResetPwdToken, userdoc.ID)
				ts.EXPECT().Create(
					any, models.ResetPwdToken, userdoc.ID,
				).Return(token, nil)
				m.EXPECT().Send(any, "Forgotten Password", any).MaxTimes(1)
				m.EXPECT().URL().MaxTimes(1)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())
		})
	}
}

func testForgottenPasswordConfirm(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	tt := []struct {
		err    error
		status int
	}{
		{err: nil, status: http.StatusOK},
		{err: errors.New("invalid token"), status: http.StatusBadRequest},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			mock := gomock.NewController(t)
			defer mock.Finish()

			id := uuid.New()
			ts := mocks.NewMockTokenStore(mock)
			e.TokenStore = ts
			router := New(e)

			r := httptest.NewRequest("GET",
				"/auth/forgotten_password/"+id.String(), nil)

			ts.EXPECT().Get(any, models.ResetPwdToken, id).Return(
				&models.Token{}, tc.err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())
		})
	}
}

func testForgottenPasswordChange(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	doc := stores.NewTestUser(t, e.UserStore)
	user := doc.Data.(*models.User)
	router := New(e)

	token, err := e.TokenStore.Create(ctx, models.ResetPwdToken, doc.ID)
	if err != nil {
		t.Fatalf("Can't create token: %s", err)
	}

	loginWith(t, router, user.Email, "foo", httptest.NewRequest("", "/", nil))

	t.Run("invalidate fail", func(t *testing.T) {
		mock := gomock.NewController(t)
		defer mock.Finish()

		ts := mocks.NewMockTokenStore(mock)
		mockedEnv := *e
		mockedEnv.TokenStore = ts
		mockedRouter := New(&mockedEnv)
		ts.EXPECT().Get(any, models.ResetPwdToken, token.ID).Return(token, nil)
		ts.EXPECT().Invalidate(any, models.ResetPwdToken, token.ID).Return(
			errors.New("something went wrong"))

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST",
			"/auth/forgotten_password/"+token.ID.String(),
			strings.NewReader("bar"),
		)
		mockedRouter.ServeHTTP(w, r)
		compareRespCode(t, http.StatusInternalServerError, w.Code, w.Body.String())

		// login creates a token.
		ts.EXPECT().Create(any, any, any).Times(1).Return(
			models.NewToken(models.SessionToken, uuid.New()), nil)
		loginWith(t, mockedRouter, user.Email, "foo", httptest.NewRequest("", "/", nil))
	})

	t.Run("valid", func(t *testing.T) {
		r := httptest.NewRequest("POST",
			"/auth/forgotten_password/"+token.ID.String(),
			strings.NewReader("bar"),
		)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
		loginWith(t, router, user.Email, "bar", httptest.NewRequest("", "/", nil))

		t.Run("invalid token on repeat", func(t *testing.T) {
			r := httptest.NewRequest("POST",
				"/auth/forgotten_password/"+token.ID.String(),
				strings.NewReader("baz"),
			)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, http.StatusBadRequest, w.Code, w.Body.String())
			loginWith(t, router, user.Email, "bar", httptest.NewRequest("", "/", nil))
		})
	})
}
