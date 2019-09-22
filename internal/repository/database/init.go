package database

import (
	"log"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/inmemorydb"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/mysqldb"
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
	GetRepository().Close()
	SetRepository(nil)
}

func GetRepository() Repository {
	if ActiveRepository == nil {
		log.Fatal("You must Open() the database before using it. This is an error in your implementation.")
	}
	return ActiveRepository
}
