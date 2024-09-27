package application

import (
	"fmt"
	"github.com/spartanhooah/testing-rest-api/data"
	"github.com/spartanhooah/testing-rest-api/db/repository"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestApplication_generateTokenPair(t *testing.T) {
	type fields struct {
		Datasource string
		DB         repository.DatabaseRepo
		Domain     string
		JWTSecret  string
	}
	type args struct {
		user *data.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    TokenPairs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &Application{
				Datasource: tt.fields.Datasource,
				DB:         tt.fields.DB,
				Domain:     tt.fields.Domain,
				JWTSecret:  tt.fields.JWTSecret,
			}
			got, err := app.generateTokenPair(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateTokenPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateTokenPair() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplication_getTokenFromHeaderAndVerify(t *testing.T) {
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	tests := []struct {
		name          string
		token         string
		errorExpected bool
		setHeader     bool
		issuer        string
	}{
		{"valid", fmt.Sprintf("Bearer %s", tokens.AccessToken), false, true, app.Domain},
		{"valid expired", fmt.Sprintf("Bearer %s", expiredToken), true, true, app.Domain},
		{"no header", "", true, false, app.Domain},
		{"invalid token", fmt.Sprintf("Bearer %s1", tokens.AccessToken), true, true, app.Domain},
		{"no bearer", fmt.Sprintf("Bear %s", tokens.AccessToken), true, true, app.Domain},
		{"three header parts", fmt.Sprintf("Bearer %s world", tokens.AccessToken), true, true, app.Domain},
		// this test must be the last one
		{"wrong issuer", fmt.Sprintf("Bearer %s", tokens.AccessToken), true, true, "anotherdomain.com"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.issuer != app.Domain {
				app.Domain = test.issuer
				tokens, _ = app.generateTokenPair(&testUser)
			}

			req, _ := http.NewRequest(http.MethodGet, "/", nil)

			if test.setHeader {
				req.Header.Add("Authorization", test.token)
			}

			resp := httptest.NewRecorder()

			_, _, err := app.getTokenFromHeaderAndVerify(resp, req)

			if err != nil && !test.errorExpected {
				t.Errorf("%s did not expect error, but got one: %s", test.name, err.Error())
			}

			if err == nil && test.errorExpected {
				t.Errorf("%s expected error, but got none", test.name)
			}

			app.Domain = "example.com"
		})
	}
}
