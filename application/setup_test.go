package application

import (
	"github.com/spartanhooah/testing-rest-api/db/repository/dbrepo"
	"os"
	"testing"
)

var app Application
var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXVkIjoiZXhhbXBsZS5jb20iLCJleHAiOjE3MjY5MjEzMDUsImlzcyI6ImV4YW1wbGUuY29tIiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMSJ9.ll69Pf3x0BLd4tzoupIwFm4W1PQLXbeMszfFqKixTSw"

func TestMain(m *testing.M) {
	app.DB = &dbrepo.TestDBRepo{}
	app.Domain = "example.com"
	app.JWTSecret = "super-secret"

	os.Exit(m.Run())
}
