package database

import (
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/historizeddb"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/inmemorydb"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/mysqldb"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
)

var (
	ActiveRepository dbrepo.Repository
)

// only exported so you can use it in test code - use Open()
func SetRepository(repository dbrepo.Repository) {
	ActiveRepository = repository
}

func Open() {
	var r dbrepo.Repository
	if config.DatabaseUse() == "mysql" {
		logging.NoCtx().Info("Opening mysql database...")
		r = historizeddb.Create(mysqldb.Create())
	} else {
		logging.NoCtx().Info("Opening inmemory database...")
		r = historizeddb.Create(inmemorydb.Create())
	}
	r.Open()
	SetRepository(r)
}

func Close() {
	logging.NoCtx().Info("Closing database...")
	GetRepository().Close()
	SetRepository(nil)
}

func MigrateIfSwitchedOn() {
	if config.MigrateDatabase() {
		logging.NoCtx().Info("Migrating database...")
		GetRepository().Migrate()
	} else {
		logging.NoCtx().Info("Not migrating database. Provide -migrate-database command line switch to enable.")
	}
}

func GetRepository() dbrepo.Repository {
	if ActiveRepository == nil {
		logging.NoCtx().Fatal("You must Open() the database before using it. This is an error in your implementation.")
	}
	return ActiveRepository
}
