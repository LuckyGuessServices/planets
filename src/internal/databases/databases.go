package databases

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/DATA-DOG/go-txdb"
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

var dbMain = struct {
	dbMutex sync.Mutex
	dbPool  *sql.DB
	// Used for [shutdown_cleanup.Register] and [shutdown_cleanup.Unregister].
	shutdownFuncId            string
	replaceNormalPoolWithTxDB bool
}{
	dbPool:                    nil,
	shutdownFuncId:            "DB: close Main pool",
	replaceNormalPoolWithTxDB: false,
}

// Main returns the same main database connections pool. If the pool has not been already created,
// then firstly creates it and stores in a local variable for future function calls.
// The pool is closed automatically during the application shutdown (see [shutdown_cleanup.Register]).
func Main() *sql.DB {
	dbMain.dbMutex.Lock()
	defer dbMain.dbMutex.Unlock()

	if nil != dbMain.dbPool {
		return dbMain.dbPool
	}

	envConfig := env.Config()
	dbConfig := NewDBMainConfig()
	dbConfig.DBName = envConfig.DBMainDatabaseName()
	dbConfig.Host = envConfig.DBMainHost()
	dbConfig.Port = envConfig.DBMainPort()
	dbConfig.UserName = envConfig.DBMainUsername()
	dbConfig.UserPassword = envConfig.DBMainPassword()
	dbConfigDebugInfo := dbConfig.DebugInfo()

	if dbMain.replaceNormalPoolWithTxDB {
		dbMain.dbPool = openTxDB(DBMainDriverName, dbConfig.DSN(), dbConfigDebugInfo)
	} else {
		dbMain.dbPool = OpenAndPing(DBMainDriverName, dbConfig.DSN(), dbConfigDebugInfo)
	}

	if !dbMain.replaceNormalPoolWithTxDB && !shutdown_cleanup.IsRegistered(dbMain.shutdownFuncId) {
		shutdown_cleanup.Register(dbMain.shutdownFuncId, func() {
			wasOpened, err := CloseMain()
			if err != nil {
				luglog.Print("[DB] Failed to close Main pool: ", err)
			} else if !wasOpened {
				luglog.Print("[DB] Main pool was closed earlier.")
			} else {
				luglog.Print("[DB] Main pool is closed.")
			}
		})
	}
	if !dbMain.replaceNormalPoolWithTxDB {
		luglog.Print("[DB] Main pool is initialized: ", dbConfigDebugInfo)
	}

	return dbMain.dbPool
}

// CloseMain closes the main db pool. The internal variable holding a pointer to the pool is emptied (`nil`)
// even if [sql.DB.Close] returns an error.
func CloseMain() (wasOpened bool, err error) {
	dbMain.dbMutex.Lock()
	defer dbMain.dbMutex.Unlock()

	if nil == dbMain.dbPool {
		return false, nil
	}

	err = dbMain.dbPool.Close()
	dbMain.dbPool = nil

	return true, err
}

func OpenAndPing(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	db := open(driverName, dataSourceName, debugInfo)
	ping(db, driverName, debugInfo)

	return db
}

func openTxDB(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	env.PanicIfEnvNotTest()

	dbTxDB := sql.OpenDB(txdb.New(driverName, dataSourceName))
	ping(dbTxDB, fmt.Sprintf("%s(TxDB)", driverName), debugInfo)

	return dbTxDB
}

func open(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		luglog.Fatalf("Unable to open '%s' database. Error: '%v'; DSN: '%s'", driverName, err, debugInfo)
	}

	return db
}

func ping(db *sql.DB, driverName string, debugInfo string) {
	if err := db.Ping(); err != nil {
		_ = db.Close()
		luglog.Fatalf("Unable to ping '%s' database. Error: '%v'; DSN: '%s'", driverName, err, debugInfo)
	}
}

// ReplaceMainWithTxDB closes the current db pool stored in an unexported global variable and enables [Main]
// to init a special db pool for tests on the next call.
// Then [Main] and other "db main pool" related functions will return this special pool.
//
// TxDB pool ensures all queries are made within a single transaction, which is always rolled back as soon as
// the pool is closed.
//
// You have to ensure this TxDB pool is closed before each test starts (or after each test ends) its work.
func ReplaceMainWithTxDB() {
	dbMain.dbMutex.Lock()
	defer dbMain.dbMutex.Unlock()

	dbMain.replaceNormalPoolWithTxDB = true

	closeMainFunc, err := shutdown_cleanup.Unregister(dbMain.shutdownFuncId)
	if err != nil {
		luglog.Fatalf("Failed to unregister '%s': %v", dbMain.shutdownFuncId, err)
	}
	closeMainFunc()
}
