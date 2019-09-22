package database

import (
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/inmemorydb"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/mysqldb"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
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
		logging.NoCtx().Info("Opening mysql database...")
		r = Repository(&mysqldb.MysqlRepository{})
	} else {
		logging.NoCtx().Info("Opening inmemory database...")
		r = Repository(&inmemorydb.InMemoryRepository{})
	}
	r.Open()
	SetRepository(r)
}

func Close() {
	logging.NoCtx().Info("Closing database...")
	GetRepository().Close()
	SetRepository(nil)
}

func Migrate() {
	// TODO make this depend on a cmd line switch. Way too dangerous otherwise.
	logging.NoCtx().Info("Migrating database...")
	GetRepository().Migrate()
}

func GetRepository() Repository {
	if ActiveRepository == nil {
		logging.NoCtx().Fatal("You must Open() the database before using it. This is an error in your implementation.")
	}
	return ActiveRepository
}
