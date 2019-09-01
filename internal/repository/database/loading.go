package database

import (
	"log"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/internal/repository/database/inmemorydb"
)

var (
	ActiveRepository Repository
)

func SetRepository(repository Repository) {
	ActiveRepository = repository
}

func Initialize() {
	if config.DatabaseUse() == "mysql" {
		log.Fatal("mysql repo not implemented yet")
	} else {
		SetRepository(&inmemorydb.InMemoryRepository{})
	}
}

func GetRepository() Repository {
	return ActiveRepository
}
