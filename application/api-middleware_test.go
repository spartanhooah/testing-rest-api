package application

import (
	"fmt"
	"github.com/spartanhooah/testing-rest-api/data"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_app_enableCORS(t *testing.T) {
	nextHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {})

	tests := []struct {
		name         string
		method       string
		expectHeader bool
	}{
		{"preflight", "OPTIONS", true},
		{"get", "GET", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handlerToTest := app.enableCORS(nextHandler)
			req := httptest.NewRequest(test.method, "/", nil)
			resp := httptest.NewRecorder()

			handlerToTest.ServeHTTP(resp, req)

			if test.expectHeader && resp.Header().Get("Access-Control-Allow-Credentials") == "" {
				t.Errorf("%s expected the CORS header but didn't find it", test.name)
			}

			if !test.expectHeader && resp.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Errorf("%s expected no CORS header but found it", test.name)
			}
		})
	}
}

func Test_app_authRequired(t *testing.T) {
	nextHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {})
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	tests := []struct {
		name             string
		token            string
		expectAuthorized bool
		setHeader        bool
	}{
		{"valid token", fmt.Sprintf("Bearer %s", tokens.AccessToken), true, true},
		{"no token", "", false, false},
		{"invalid token", fmt.Sprintf("Bearer %s", expiredToken), false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)

			if test.setHeader {
				req.Header.Set("Authorization", test.token)
			}

			resp := httptest.NewRecorder()
			handlerToTest := app.authRequired(nextHandler)

			handlerToTest.ServeHTTP(resp, req)

			if test.expectAuthorized && resp.Code == http.StatusUnauthorized {
				t.Errorf("%s got status 401 but should not have", test.name)
			}

			if !test.expectAuthorized && resp.Code != http.StatusUnauthorized {
				t.Errorf("%s got status 401 but should have", test.name)
			}
		})
	}
}
