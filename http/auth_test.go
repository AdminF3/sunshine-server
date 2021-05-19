package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stageai.tech/sunshine/sunshine/models"
	"stageai.tech/sunshine/sunshine/services"
	"stageai.tech/sunshine/sunshine/stores"

	"github.com/google/uuid"
)

func TestAuth(t *testing.T) {
	t.Run("login", testLogin)
	t.Run("change_password", testChangePassword)
	t.Run("session", testSessions)
	t.Run("session/bad", testBadSessions)
}

func testLogin(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	u := stores.NewTestUser(t, e.UserStore)

	tt := []struct {
		name   string
		body   string
		status int
	}{
		{
			name:   "valid",
			body:   fmt.Sprintf(`{"email":"%s", "password": "foo"}`, u.Data.(*models.User).Email),
			status: http.StatusOK,
		},
		{
			name:   "wrong password",
			body:   fmt.Sprintf(`{"email":"%s", "password": "bar"}`, u.Data.(*models.User).Email),
			status: http.StatusUnauthorized,
		},
		{
			name:   "wrong email",
			body:   `{"email":"not_john_doe@example.com", "password": "bar"}`,
			status: http.StatusUnauthorized,
		},
		{
			name:   "empty body",
			body:   ``,
			status: http.StatusBadRequest,
		},
		{
			name:   "malformed json",
			body:   `{"email":"john_doe@example.com"`,
			status: http.StatusBadRequest,
		},
		{
			name:   "missing email",
			body:   `{"password": "foo"}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "missing password",
			body:   `{"email": "john_doe@example.com"}`,
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/auth/login",
				strings.NewReader(tc.body))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			var validCookies []http.Cookie
			for _, c := range w.Result().Cookies() {
				if c.MaxAge > 0 {
					validCookies = append(validCookies, *c)
				}
			}
			if (len(validCookies) > 0) != (w.Code == http.StatusOK) {
				t.Errorf("Expected valid cookie *only* on success; got %#v on %d",
					validCookies, w.Code)
			}

			if w.Code == http.StatusOK {
				doc := models.Document{Data: &models.User{}}
				err := json.NewDecoder(w.Body).Decode(&doc)
				if err != nil {
					t.Fatalf("can't decode user from success login: %s", err)
				}
			}

		})
	}
}

func testChangePassword(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	tt := []struct {
		name   string
		body   string
		status int
		unauth bool
	}{
		{
			name:   "valid",
			body:   `{"old": "foo", "new": "bar"}`,
			status: http.StatusOK,
		},
		{
			name:   "not logged in",
			body:   `{"old": "foo", "new": "bar"}`,
			status: http.StatusUnauthorized,
			unauth: true,
		},
		{
			name:   "wrong password",
			body:   `{"old": "wrong", "new": "bar"}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid new",
			body:   `{"old": "foo", "new": ""}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "empty body",
			body:   ``,
			status: http.StatusBadRequest,
		},
		{
			name:   "malformed json",
			body:   `{"old":"foo"`,
			status: http.StatusBadRequest,
		},
		{
			name:   "missing old",
			body:   `{"new": "bar"}`,
			status: http.StatusBadRequest,
		},
	}

	deluser := func(id uuid.UUID) {
		d, err := e.UserStore.Get(ctx, id)
		if err != nil {
			t.Fatalf("get for deluser: %s", err)
		}
		err = e.UserStore.Delete(context.Background(), d)
		if err != nil {
			t.Fatalf("deluser: %s", err)
		}
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			user := stores.NewTestUser(t, e.UserStore)
			defer deluser(user.ID)

			r := httptest.NewRequest(
				"POST",
				"/auth/change_password",
				strings.NewReader(tc.body))
			if !tc.unauth {
				r = loginAs(t, e, user, r)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			compareRespCode(t, tc.status, w.Code, w.Body.String())

			if w.Code == http.StatusOK {
				loginWith(t, router, user.Data.(*models.User).Email, "bar", r)
			} else {
				loginWith(t, router, user.Data.(*models.User).Email, "foo", r)
			}

		})
	}
}

func testSessions(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	user := stores.NewTestUser(t, e.UserStore).Data.(*models.User)

	// login
	loginPayload := fmt.Sprintf(`{"email":"%s", "password": "foo"}`, user.Email)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/auth/login",
		strings.NewReader(loginPayload))
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	cookie := w.Header().Get("Set-Cookie")

	// update user
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/user/"+user.ID.String(),
		strings.NewReader(`{"name": "Logged In"}`))
	r.Header.Add("Cookie", cookie)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	udoc, err := e.UserStore.Get(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if name := udoc.Data.(*models.User).Name; name != "Logged In" {
		t.Fatalf("Update with valid session set the name to: %q", name)
	}

	// logout
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/auth/logout", nil)
	r.Header.Add("Cookie", cookie)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())

	// update user again reusing the old cookie
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/user/"+user.ID.String(),
		strings.NewReader(`{"name": "Logged Out"}`))
	r.Header.Add("Cookie", cookie)
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())
	udoc, err = e.UserStore.Get(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if name := udoc.Data.(*models.User).Name; name != "Logged In" {
		t.Fatalf("Update with stale session set the name to: %q", name)
	}
}

// testBadSessions makes sure we don't crash on malicious session payload.
func testBadSessions(t *testing.T) {
	e, del := newTestEnv(t)
	defer del()

	router := New(e)
	user := stores.NewTestUser(t, e.UserStore).Data.(*models.User)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/user/"+user.ID.String(),
		strings.NewReader(`{"name": "hax0r"}`))
	r.Header.Add("Cookie", user.ID.String())
	router.ServeHTTP(w, r)
	compareRespCode(t, http.StatusUnauthorized, w.Code, w.Body.String())
}

func loginWith(t *testing.T, router http.Handler, email, password string, r *http.Request) *http.Request {
	l := login{Email: email, Password: password}
	lb, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("loginAs failed on json.Marshal: %s", err)
	}

	lr := httptest.NewRequest("POST", "/auth/login",
		strings.NewReader(string(lb)))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, lr)
	compareRespCode(t, http.StatusOK, w.Code, w.Body.String())
	r.Header.Add("Cookie", w.Header().Get("Set-Cookie"))
	return r
}

func loginAs(t *testing.T, e *services.Env, user *models.Document, r *http.Request) *http.Request {
	if user == nil {
		return r
	}
	ctx := services.NewTestContext(t, e, user)

	w := httptest.NewRecorder()
	s := services.Session(e.SessionStore, r)
	s.Values["id"] = services.FromContext(ctx).ID
	s.Values["uuid"] = user.ID
	services.SaveSession(s, r, w)
	r.Header.Add("Cookie", w.Header().Get("Set-Cookie"))
	return r
}
