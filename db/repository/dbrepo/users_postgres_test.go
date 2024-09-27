//go:build integration

// the above allows separating some tests; can run just integration tests with -tags=integration

package dbrepo

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spartanhooah/profile-picture-web/data"
	"github.com/spartanhooah/profile-picture-web/db/repository"
	"log"
	"os"
	"testing"
	"time"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

// sets up test environment
func TestMain(m *testing.M) {
	// connect to Docker; fail if Docker not present
	p, err := dockertest.NewPool("")

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	pool = p

	// set up Docker options, specifying image, etc.
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	// get a resource (image)
	resource, err = pool.RunWithOptions(&opts)

	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not start resource: %s", err)
	}

	// start image, wait until it's ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, user, password, dbName, port))

		if err != nil {
			log.Println("Error:", err)
			return err
		}

		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to database: %s", err)
	}

	// populate DB with empty tables
	err = createTables()

	if err != nil {
		log.Fatalf("Error creating tables: %s", err)
	}

	testRepo = &PostgresDBRepo{DB: testDB}

	// run the tests
	code := m.Run()

	// clean up
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Error purging resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")

	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()

	if err != nil {
		t.Error("Can't ping DB")
	}
}

func Test_PostgresDBRepo_InsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	id, err := testRepo.InsertUser(testUser)

	if err != nil {
		t.Errorf("Error inserting user: %s", err)
	}

	if id != 1 {
		t.Errorf("Incorrect id returned; expected 1 but got %d", id)
	}
}

func Test_PostgresDBRepo_GetAllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()

	if err != nil {
		t.Errorf("Error getting all users: %s", err)
	}

	if len(users) != 1 {
		t.Errorf("Incorrect number of users returned; expected 1 but got %d", len(users))
	}

	testUser := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "jack@smith.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()

	if err != nil {
		t.Errorf("Error getting all users: %s", err)
	}

	if len(users) != 2 {
		t.Errorf("Incorrect number of users returned; expected 2 but got %d", len(users))
	}
}

func Test_PostgresDBRepo_GetUserById(t *testing.T) {
	user, err := testRepo.GetUser(1)

	if err != nil {
		t.Errorf("Error getting user: %s", err)
	}

	if user.Email != "admin@example.com" {
		t.Errorf("Incorrect email returned; expected admin@example.com but got %s", user.Email)
	}
}

func Test_PostgresDBRepo_GetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("jack@smith.com")

	if err != nil {
		t.Errorf("Error getting user: %s", err)
	}

	if user.ID != 2 {
		t.Errorf("Incorrect id returned; expected 2 but got %d", user.ID)
	}
}

func Test_PostgresDBRepo_UpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(2)

	user.FirstName = "Jane"
	user.Email = "jane@example.com"

	err := testRepo.UpdateUser(*user)

	if err != nil {
		t.Errorf("Error updating user: %s", err)
	}

	if user.FirstName != "Jane" {
		t.Errorf("Incorrect first name returned; expected Jane but got %s", user.FirstName)
	}

	if user.Email != "jane@example.com" {
		t.Errorf("Incorrect email returned; expected jane@example.com but got %s", user.Email)
	}
}

func Test_PostgresDBRepo_DeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)

	if err != nil {
		t.Errorf("Error deleting user: %s", err)
	}

	_, err = testRepo.GetUser(2)

	if err == nil {
		t.Errorf("Found user 2 but that user should have been deleted")
	}
}

func Test_PostgresDBRepo_ResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "password")

	if err != nil {
		t.Errorf("Error resetting password: %s", err)
	}

	user, _ := testRepo.GetUser(1)

	matches, err := user.PasswordMatches("password")

	if err != nil {
		t.Errorf("Error verifying password: %s", err)
	}

	if !matches {
		t.Errorf("Password did not match")
	}
}

func Test_PostgresDBRepo_InsertUserImage(t *testing.T) {
	image := data.UserImage{
		UserID:    1,
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	imageId, err := testRepo.InsertUserImage(image)

	if err != nil {
		t.Errorf("Error inserting image: %s", err)
	}

	if imageId != 1 {
		t.Errorf("Incorrect id returned; expected 1 but got %d", imageId)
	}

	image.UserID = 100

	_, err = testRepo.InsertUserImage(image)

	if err == nil {
		t.Errorf("Should not have been able attach image to nonexistent user")
	}
}
