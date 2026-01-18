package databases

import (
	"database/sql"
	"fmt"

	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"

	_ "github.com/lib/pq"
)

const DBMainDriverName = "postgres"

type DBMainConfig struct {
	ClientEncoding string
	ConnectTimeout int
	DBName         string
	Host           string
	Port           int
	SSLMode        string
	TimeZone       string
	UserName       string
	UserPassword   string
}

func (config *DBMainConfig) String() string {
	return config.DebugInfo()
}

func (config *DBMainConfig) DSN() string {
	return config.dsnInternal(false)
}

func (config *DBMainConfig) DebugInfo() string {
	return config.dsnInternal(true)
}

func (config *DBMainConfig) dsnInternal(hidePassword bool) string {
	var password string
	if hidePassword {
		password = "(hidden)"
	} else {
		password = config.UserPassword
	}

	return fmt.Sprintf(
		"client_encoding=%s connect_timeout=%d dbname=%s host=%s port=%d sslmode=%s TimeZone=%s user=%s password=%s",
		config.ClientEncoding,
		config.ConnectTimeout,
		config.DBName,
		config.Host,
		config.Port,
		config.SSLMode,
		config.TimeZone,
		config.UserName,
		password,
	)
}

func NewDBMainConfig() *DBMainConfig {
	return &DBMainConfig{
		ClientEncoding: "UTF8",
		ConnectTimeout: 2,
		SSLMode:        "disable",
		TimeZone:       "UTC",
	}
}

var dbMain *sql.DB

// Main returns the same main database connections pool. If the pool has not been already created,
// then firstly creates it and stores in a local variable for future function calls.
// The pool is closed automatically during the application shutdown (see [shutdown_cleanup.Register]).
func Main() *sql.DB {
	if nil != dbMain {
		return dbMain
	}

	envConfig := env.Config()
	dbConfig := NewDBMainConfig()
	dbConfig.DBName = envConfig.DBMainDatabaseName()
	dbConfig.Host = envConfig.DBMainHost()
	dbConfig.Port = envConfig.DBMainPort()
	dbConfig.UserName = envConfig.DBMainUsername()
	dbConfig.UserPassword = envConfig.DBMainPassword()

	dbMain = Open(DBMainDriverName, dbConfig.DSN(), dbConfig.DebugInfo())
	shutdown_cleanup.Register("DB: close Main pool", func() {
		wasOpened, err := CloseMain()
		if err != nil {
			luglog.Print("[DB] Failed to close Main pool: ", err)
		} else if !wasOpened {
			luglog.Print("[DB] Main pool was closed earlier.")
		} else {
			luglog.Print("[DB] Main pool is closed.")
		}
	})
	luglog.Print("[DB] Main pool is initialized: ", dbConfig.DebugInfo())

	return dbMain
}

func CloseMain() (wasOpened bool, err error) {
	if nil == dbMain {
		return false, nil
	}

	err = dbMain.Close()
	dbMain = nil

	return true, err
}

func Open(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		luglog.Fatalf("Unable to open '%s' connection. Error: '%v'; DSN: '%s'", driverName, err, debugInfo)
	}

	err = db.Ping()
	if err != nil {
		_ = db.Close()
		luglog.Fatalf("Unable to ping '%s' connection. Error: '%v'; DSN: '%s'", driverName, err, debugInfo)
	}

	return db
}
