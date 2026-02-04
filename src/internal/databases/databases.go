package databases

import (
	"sync"

	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
	"github.com/uptrace/bun"
)

type dbPool struct {
	mutex sync.Mutex

	poolORM *bun.DB

	shutdownFuncId string // Used for [shutdown_cleanup.Register] and [shutdown_cleanup.Unregister].
}

var (
	dbCore                       = &dbPool{}
	dbCoreReadonly               = &dbPool{}
	replaceWriteablePoolWithTxDB = false
)

// Core returns the same core database connections pool. If the pool has not been already created,
// then firstly creates it and stores in a local variable for future function calls.
//
// Normally, the pool is closed automatically during the application shutdown (see [shutdown_cleanup.Register]).
// During tests, the pool is intended to be closed at the end of each test (see [ReplaceWritablePoolsWithTxDB]).
func Core() *bun.DB {
	return coreInternal(dbCore, false)
}

// CoreReadonly works the same as [Core], except the returned pool is connected to a (presumably) read-only database.
func CoreReadonly() *bun.DB {
	return coreInternal(dbCoreReadonly, true)
}

func coreInternal(props *dbPool, isReadonly bool) *bun.DB {
	props.mutex.Lock()
	defer props.mutex.Unlock()

	if nil != props.poolORM {
		return props.poolORM
	}

	dbConfig := NewDBPrimaryConfig()
	// For now, there are the same connection settings for master and readonly slave databases.
	// If needed in the future, these should be replaced with settings according to [isReadonly] value.
	props.poolORM = OpenAndPingPrimaryORM(dbConfig, replaceWriteablePoolWithTxDB)

	var (
		poolNameForLogs string
		poolCloseFunc   func() (bool, error)
	)
	if isReadonly {
		poolNameForLogs = "CORE readonly pool"
		poolCloseFunc = CloseCoreReadonly
	} else {
		poolNameForLogs = "CORE pool"
		poolCloseFunc = CloseCore
	}

	props.shutdownFuncId = "DB: close " + poolNameForLogs
	if !replaceWriteablePoolWithTxDB && !shutdown_cleanup.IsRegistered(props.shutdownFuncId) {
		errRegister := shutdown_cleanup.Register(props.shutdownFuncId, func() {
			wasOpened, err := poolCloseFunc()
			if err != nil {
				luglog.Printf("[DB] Failed to close %s: %v", poolNameForLogs, err)
			} else if !wasOpened {
				luglog.Printf("[DB] %s was closed earlier.", poolNameForLogs)
			} else {
				luglog.Printf("[DB] %s is closed.", poolNameForLogs)
			}
		})
		if errRegister != nil {
			luglog.Panicf("[DB] Failed to register %s shutdown function: %v", poolNameForLogs, errRegister)
		}
	}
	if !replaceWriteablePoolWithTxDB {
		luglog.Printf("[DB] %s is initialized: %s", poolNameForLogs, dbConfig.DSNForLogs())
	}

	return props.poolORM
}

// CloseCore closes the core db (ORM) pool.
// The internal variable holding pointer to a pool is emptied (`nil`) even if [sql.DB.Close] returns an error.
//
// Returns 1) if a pool was actually opened (before calling this function) and 2) the operation error.
func CloseCore() (bool, error) {
	return closeCoreInternal(dbCore)
}

// CloseCoreReadonly does the same as [CloseCore] except closing the readonly pool.
func CloseCoreReadonly() (bool, error) {
	return closeCoreInternal(dbCoreReadonly)
}

func closeCoreInternal(props *dbPool) (bool, error) {
	props.mutex.Lock()
	defer props.mutex.Unlock()

	if nil == props.poolORM {
		return false, nil
	}

	errClose := props.poolORM.Close()
	props.poolORM = nil

	return true, errClose
}

// ReplaceWritablePoolsWithTxDB closes current "constant" db pools stored in unexported global variables and enables
// future pools inits to open special db pools for tests on the next call.
// Then [Core] and other "core db pool" related functions will return this special pool.
//
// TxDB pool ensures all queries are made within a single transaction, which is always rolled back as soon as
// the pool is closed.
//
// You have to ensure this TxDB pool is closed before each test starts (or after each test ends) its work.
//
// The function is intended to be called once before all tests launch. So there should be no concurrency.
// In case any concurrency happens, it happens only because a development oversight. One of the cases described
// below will happen and indicate the problem. Then the problem should be dealt with outside of this function scope.
//
// 1. A concurrent request receives a pointer to a not-yet closed pool. Then this function closes that pool. Then the
// next query made via that pool fails.
//
// 2. A concurrent request calls this function too. When both goroutines try to unregister a shutdown function, one of
// goroutines will do it later and eventually produce a panic because of failing to unregister a function.
func ReplaceWritablePoolsWithTxDB() {
	env.PanicIfEnvNotTest()

	if replaceWriteablePoolWithTxDB {
		luglog.Panic("[DB] TxDB driver was requested already.")
	}
	replaceWriteablePoolWithTxDB = true

	if dbCore.poolORM != nil {
		closeFunc, errUnregister := shutdown_cleanup.Unregister(dbCore.shutdownFuncId)
		if errUnregister != nil {
			luglog.Panicf("Failed to unregister '%s': %v", dbCore.shutdownFuncId, errUnregister)
		}
		closeFunc()
	}

	if dbCoreReadonly.poolORM != nil {
		closeFunc, errUnregister := shutdown_cleanup.Unregister(dbCoreReadonly.shutdownFuncId)
		if errUnregister != nil {
			luglog.Panicf("Failed to unregister '%s': %v", dbCoreReadonly.shutdownFuncId, errUnregister)
		}
		closeFunc()
	}
}
