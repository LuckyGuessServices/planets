package databases

import (
	"database/sql"
	"sync"

	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
)

type dbMainType struct {
	mutex                     sync.Mutex
	pool                      *sql.DB
	replaceNormalPoolWithTxDB bool

	// Used for [shutdown_cleanup.Register] and [shutdown_cleanup.Unregister].
	shutdownFuncId string
}

var dbMain = dbMainType{
	pool:                      nil,
	replaceNormalPoolWithTxDB: false,
}

// Main returns the same main database connections pool. If the pool has not been already created,
// then firstly creates it and stores in a local variable for future function calls.
// The pool is closed automatically during the application shutdown (see [shutdown_cleanup.Register]).
func Main() *sql.DB {
	dbMain.mutex.Lock()
	defer dbMain.mutex.Unlock()

	if nil != dbMain.pool {
		return dbMain.pool
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
		dbMain.pool = OpenAndPingTxDB(DBMainDriverName, dbConfig.DSN(), dbConfigDebugInfo)
	} else {
		dbMain.pool = OpenAndPing(DBMainDriverName, dbConfig.DSN(), dbConfigDebugInfo)
	}

	dbMain.shutdownFuncId = "DB: close Main pool"
	if !dbMain.replaceNormalPoolWithTxDB && !shutdown_cleanup.IsRegistered(dbMain.shutdownFuncId) {
		errRegister := shutdown_cleanup.Register(dbMain.shutdownFuncId, func() {
			wasOpened, err := CloseMain()
			if err != nil {
				luglog.Print("[DB] Failed to close Main pool: ", err)
			} else if !wasOpened {
				luglog.Print("[DB] Main pool was closed earlier.")
			} else {
				luglog.Print("[DB] Main pool is closed.")
			}
		})
		if errRegister != nil {
			luglog.Panic("[DB] Failed to register Main pool shutdown function: ", errRegister)
		}
	}
	if !dbMain.replaceNormalPoolWithTxDB {
		luglog.Print("[DB] Main pool is initialized: ", dbConfigDebugInfo)
	}

	return dbMain.pool
}

// CloseMain closes the main db pool. The internal variable holding a pointer to the pool is emptied (`nil`)
// even if [sql.DB.Close] returns an error.
func CloseMain() (bool, error) {
	dbMain.mutex.Lock()
	defer dbMain.mutex.Unlock()

	if nil == dbMain.pool {
		return false, nil
	}

	err := dbMain.pool.Close()
	dbMain.pool = nil

	return true, err
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
	env.PanicIfEnvNotTest()

	// The function is intended to be called once before all tests launch. So there should be no concurrency.
	// But if concurrency actually happens, let's rely on a mutex... However, unlock it as soon as necessary checks
	// are made. If unlocking too late, then closing the original database pool will deadlock (requests the same mutex).
	func() {
		dbMain.mutex.Lock()
		defer dbMain.mutex.Unlock()

		if dbMain.replaceNormalPoolWithTxDB {
			luglog.Panic("[DB] TxDB driver was requested already.")
		}
		dbMain.replaceNormalPoolWithTxDB = true

		// Stop here, if for some reason the pool has been already closed.
		if dbMain.pool == nil {
			return
		}
	}()

	closeMainFunc, errUnregister := shutdown_cleanup.Unregister(dbMain.shutdownFuncId)
	if errUnregister != nil {
		luglog.Panicf("Failed to unregister '%s': %v", dbMain.shutdownFuncId, errUnregister)
	}
	closeMainFunc()
}
