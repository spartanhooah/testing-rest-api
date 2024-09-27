package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/spartanhooah/testing-rest-api/application"
	"github.com/spartanhooah/testing-rest-api/db/repository/dbrepo"
	"log"
	"net/http"
)

const port = 8090

func main() {
	var app application.Application
	flag.StringVar(&app.Datasource, "datasource", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.StringVar(&app.Domain, "domain", "example.com", "Domain for application, e.g. company.com")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "super-secret", "signing secret")
	flag.Parse()

	conn, err := app.ConnectToDB()

	if err != nil {
		log.Fatal(err)
	}

	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			log.Fatal("Error closing connection", err)
		}
	}(conn)

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	log.Printf("Starting API on port %d\n", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.Routes())

	if err != nil {
		log.Fatal(err)
	}
}
