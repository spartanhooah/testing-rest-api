package application

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/spartanhooah/testing-rest-api/data"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func Test_app_authenticate(t *testing.T) {
	var tests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
		{"not JSON", `I'm not JSON'`, http.StatusUnauthorized},
		{"empty JSON", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email":"","password":"secret"}`, http.StatusUnauthorized},
		{"empty password", `{"email":"admin@example.com","password":""}`, http.StatusUnauthorized},
		{"invalid user", `{"email":"admin@otherdomain.com","password":"secret"}`, http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.requestBody)
			req, _ := http.NewRequest("POST", "/auth", reader)
			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(app.authenticate)

			handler.ServeHTTP(resp, req)

			if test.expectedStatusCode != resp.Code {
				t.Errorf("%s expected status code %d, got %d", test.name, test.expectedStatusCode, resp.Code)
			}
		})
	}
}

func Test_app_refresh(t *testing.T) {
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid token", "", http.StatusOK, true},
		{"expired token", expiredToken, http.StatusBadRequest, true},
		{"no need to refresh", "", http.StatusTooEarly, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := RefreshTokenExpiry

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var token string

			if test.token == "" {
				if test.resetRefreshTime {
					RefreshTokenExpiry = time.Second * 1
				}

				tokens, _ := app.generateTokenPair(&testUser)
				token = tokens.RefreshToken
			} else {
				token = test.token
			}

			postedData := url.Values{
				"refresh_token": {token},
			}

			req, _ := http.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp := httptest.NewRecorder()
			handler := http.HandlerFunc(app.refresh)

			handler.ServeHTTP(resp, req)

			if test.expectedStatusCode != resp.Code {
				t.Errorf("%s expected status code %d, got %d", test.name, test.expectedStatusCode, resp.Code)
			}
		})

		RefreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_usersCRUD(t *testing.T) {
	var tests = []struct {
		name               string
		method             string
		json               string
		idParam            string
		handler            http.HandlerFunc
		expectedStatusCode int
	}{
		{"get all users", "GET", "", "", app.allUsers, http.StatusOK},
		{"delete user", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"delete user bad URL param", "DELETE", "", "n", app.deleteUser, http.StatusBadRequest},
		{"get user valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"get user invalid", "GET", "", "5", app.getUser, http.StatusBadRequest},
		{"get user bad URL param", "GET", "", "y", app.getUser, http.StatusBadRequest},
		{
			"update user valid",
			"PATCH",
			`{"id":1,"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusNoContent,
		},
		{
			"update user invalid",
			"PATCH",
			`{"id":5,"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusInternalServerError,
		},
		{
			"update user invalid JSON",
			"PATCH",
			`{"id":1,first_name:"Administrator","last_name":"User","email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"insert user valid",
			"PUT",
			`{"first_name":"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.createUser,
			http.StatusCreated,
		},
		{
			"insert user invalid",
			"PUT",
			`{"foo":"bar","first_name":"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.createUser,
			http.StatusBadRequest,
		},
		{
			"insert user invalid json",
			"PUT",
			`{first_name:"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.createUser,
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var req *http.Request

			if test.json == "" {
				req, _ = http.NewRequest(test.method, "/", nil)
			} else {
				req, _ = http.NewRequest(test.method, "/", strings.NewReader(test.json))
			}

			if test.idParam != "" {
				chiCtx := chi.NewRouteContext()
				chiCtx.URLParams.Add("userId", test.idParam)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			}

			resp := httptest.NewRecorder()
			handler := test.handler

			handler.ServeHTTP(resp, req)

			if test.expectedStatusCode != resp.Code {
				t.Errorf("%s expected status code %d, got %d", test.name, test.expectedStatusCode, resp.Code)
			}
		})
	}
}

func Test_app_refreshUsingCookie(t *testing.T) {
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	goodCookie := &http.Cookie{
		Name:     refreshCookieName,
		Path:     "/",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(RefreshTokenExpiry),
		MaxAge:   int(RefreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	badCookie := &http.Cookie{
		Name:     refreshCookieName,
		Path:     "/",
		Value:    "someBadString",
		Expires:  time.Now().Add(RefreshTokenExpiry),
		MaxAge:   int(RefreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	tests := []struct {
		name               string
		addCookie          bool
		cookie             *http.Cookie
		expectedStatusCode int
	}{
		{"valid cookie", true, goodCookie, http.StatusOK},
		{"invalid cookie", true, badCookie, http.StatusBadRequest},
		{"no cookie", false, nil, http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)

			if test.addCookie {
				req.AddCookie(test.cookie)
			}

			handler := http.HandlerFunc(app.refreshUsingCookie)

			handler.ServeHTTP(resp, req)

			if test.expectedStatusCode != resp.Code {
				t.Errorf("%s expected status code %d, got %d", test.name, test.expectedStatusCode, resp.Code)
			}
		})
	}
}

func Test_app_deleteRefreshCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	handler := http.HandlerFunc(app.deleteRefreshCookie)

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Errorf("wrong status; expected %d, got %d", http.StatusAccepted, resp.Code)
	}

	foundCookie := false

	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == refreshCookieName {
			foundCookie = true

			if cookie.Expires.After(time.Now()) {
				t.Errorf("cookie expires in the future and should not be: %v", cookie.Expires.UTC())
			}
		}
	}

	if !foundCookie {
		t.Errorf("refresh cookie not found")
	}
}
