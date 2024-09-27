package application

import (
	"github.com/spartanhooah/testing-rest-api/db/repository"
)

type Application struct {
	Datasource string
	DB         repository.DatabaseRepo
	Domain     string
	JWTSecret  string
}
