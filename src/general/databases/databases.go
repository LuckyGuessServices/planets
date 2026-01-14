package databases

import (
	"LuG/planets/general/env"
	"LuG/planets/general/luglog"
	"LuG/planets/general/shutdown_cleanup"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var dbMain *sql.DB

func DBMain() *sql.DB {
	if nil != dbMain {
		return dbMain
	}
	luglog.Print("DBMain pool is initialized.")

	dsn := func() string {
		dsnTemplate := "host=%s port=%s user=%s password=%s dbname=%s"
		if env.ApplicationEnvironmentTest == env.AppEnv() {
			return fmt.Sprintf(
				dsnTemplate,
				env.VarValue(env.VarNameDBMainHost, "127.0.0.1", false),
				env.VarValue(env.VarNameDBMainPort, "5400", false),
				env.VarValue(env.VarNameDBMainUsername, "planets_test_user", false),
				env.VarValue(env.VarNameDBMainPassword, "planets_test_pass", false),
				env.VarValue(env.VarNameDBMainDatabaseName, "planets_test", false),
			)
		}

		return fmt.Sprintf(
			dsnTemplate,
			env.VarValue(env.VarNameDBMainHost, "127.0.0.1", false),
			env.VarValue(env.VarNameDBMainPort, "5400", false),
			env.VarValue(env.VarNameDBMainUsername, "planets_user", false),
			env.VarValue(env.VarNameDBMainPassword, "planets_pass", false),
			env.VarValue(env.VarNameDBMainDatabaseName, "planets", false),
		)
	}() +
		" connect_timeout=2 sslmode=disable client_encoding=UTF8 TimeZone=UTC"

	dbMain = Open("postgres", dsn)
	shutdown_cleanup.Register("Close DBMain pool", func() {
		wasOpened, err := CloseDBMain()
		if err != nil {
			luglog.Print("Failed to close DBMain pool: ", err)
		} else if !wasOpened {
			luglog.Print("DBMain pool was closed earlier.")
		} else {
			luglog.Print("DBMain pool is closed successfully.")
		}
	})

	return dbMain
}

func CloseDBMain() (wasOpened bool, err error) {
	if nil == dbMain {
		return false, nil
	}

	err = dbMain.Close()
	dbMain = nil

	return true, err
}

func Open(driverName string, dataSourceName string) *sql.DB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		luglog.Fatalf("Unable to open '%v' connection '%v': %v", driverName, dataSourceName, err)
	}

	err = db.Ping()
	if err != nil {
		_ = db.Close()
		luglog.Fatalf("Unable to ping '%v' connection '%v': %v", driverName, dataSourceName, err)
	}

	return db
}
