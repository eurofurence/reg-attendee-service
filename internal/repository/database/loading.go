package database

import (
	"log"
	"rexis/rexis-go-attendee/internal/repository/config"
	"rexis/rexis-go-attendee/internal/repository/database/inmemorydb"
	"rexis/rexis-go-attendee/internal/repository/database/mysqldb"
)

var (
	ActiveRepository Repository
)

// only exported so you can use it in test code - use Open()
func SetRepository(repository Repository) {
	ActiveRepository = repository
}

func Open() {
	var r Repository
	if config.DatabaseUse() == "mysql" {
		log.Print("Opening mysql database...")
		r = Repository(&mysqldb.MysqlRepository{})
	} else {
		log.Print("Opening inmemory database...")
		r = Repository(&inmemorydb.InMemoryRepository{})
	}
	r.Open()
	SetRepository(r)
}

func Close() {
	log.Print("Closing database...")
	ActiveRepository.Close()
}

func GetRepository() Repository {
	return ActiveRepository
}
