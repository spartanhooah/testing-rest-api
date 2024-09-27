package application

import (
	"database/sql"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

func openDB(datasource string) (*sql.DB, error) {
	db, err := sql.Open("pgx", datasource)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *Application) ConnectToDB() (*sql.DB, error) {
	connection, err := openDB(app.Datasource)

	if err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres")

	return connection, nil
}
