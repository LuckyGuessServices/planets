package databases

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const (
	DBPrimaryDriverName            = "pgx"
	DBPrimaryMaxOpenConnections    = 25
	DBPrimaryMaxIdleConnections    = DBPrimaryMaxOpenConnections
	DBPrimaryConnectionMaxLifetime = 5 * time.Minute
)

type DBPrimaryConfig struct {
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

func (config *DBPrimaryConfig) String() string {
	return config.DSNForLogs()
}

// DSN returns a connection string for passing to database pool openers.
func (config *DBPrimaryConfig) DSN() string {
	return config.dsnInternal(false)
}

// DSNForLogs returns the same string as [DSN] except the password value.
func (config *DBPrimaryConfig) DSNForLogs() string {
	return config.dsnInternal(true)
}

func (config *DBPrimaryConfig) dsnInternal(hideSensitiveData bool) string {
	var password string
	if hideSensitiveData {
		password = "(hidden)"
	} else {
		password = config.UserPassword
	}

	generalPostgresConnectionString := fmt.Sprintf(
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

	// Automatic prepared statements for PGX should be disabled, because the project will generally suffer from those.
	// PGX config struct could be built (instead of strings concatenation) and then passed to sql.OpenDB(), but:
	// 1. The code would become more coupled with PGX implementation. General Postgres settings would mix with PGX-only
	//    settings in a single PGX config. Replacing PGX with something else would become a bit harder.
	// 2. Stringer implementation (hiding sensitive settings like a password) is still needed for logs.
	return generalPostgresConnectionString + " default_query_exec_mode=simple_protocol"
}

func NewDBPrimaryConfig() *DBPrimaryConfig {
	envConfig := env.Config()

	return &DBPrimaryConfig{
		ClientEncoding: "UTF8",
		ConnectTimeout: 2,
		DBName:         envConfig.DBCoreDatabaseName(),
		Host:           envConfig.DBCoreHost(),
		Port:           envConfig.DBCorePort(),
		SSLMode:        "disable",
		TimeZone:       "UTC",
		UserName:       envConfig.DBCoreUsername(),
		UserPassword:   envConfig.DBCorePassword(),
	}
}

func OpenAndPingPrimaryORM(dbConfig *DBPrimaryConfig, shouldWrapWithTxDB bool) *bun.DB {
	var poolDB *sql.DB
	if shouldWrapWithTxDB {
		poolDB = OpenAndPingTxDB(DBPrimaryDriverName, dbConfig.DSN(), dbConfig.DSNForLogs())
	} else {
		poolDB = OpenAndPing(DBPrimaryDriverName, dbConfig.DSN(), dbConfig.DSNForLogs())
	}
	setUpDBPrimaryPool(poolDB)

	return bun.NewDB(poolDB, pgdialect.New())
}

func OpenAndPing(driverName string, connectionString string, connectionStringForLogs string) *sql.DB {
	db, errOpen := sql.Open(driverName, connectionString)
	if errOpen != nil {
		luglog.Panicf(
			"Unable to open '%s' database. Error: '%v'; DSN: '%s'",
			driverName,
			errOpen,
			connectionStringForLogs,
		)
	}
	ping(db, driverName, connectionStringForLogs)

	return db
}

// OpenAndPingTxDB opens a modified database pool. This pool should be used in tests only.
//
// All queries made with this pool are enclosed in a single global transaction.
// This transaction is rolled back as soon as the pool is closed.
//
// Opening and commiting / roll backing transactions within this pool actually operates with safe points.
// The single global transactions stays intact and always rolls back.
func OpenAndPingTxDB(driverName string, connectionString string, connectionStringForLogs string) *sql.DB {
	env.PanicIfEnvNotTest()

	dbTxDB := sql.OpenDB(txdb.New(driverName, connectionString))
	ping(dbTxDB, driverName+"(TxDB)", connectionStringForLogs)

	return dbTxDB
}

// setUpDBPrimaryPool adds additional setting for the go standard db pool.
func setUpDBPrimaryPool(db *sql.DB) {
	db.SetMaxOpenConns(DBPrimaryMaxOpenConnections)
	db.SetMaxIdleConns(DBPrimaryMaxIdleConnections)
	db.SetConnMaxLifetime(DBPrimaryConnectionMaxLifetime)
}

func ping(db *sql.DB, driverName string, connectionStringForLogs string) {
	if err := db.Ping(); err != nil {
		_ = db.Close()
		luglog.Panicf("Unable to ping '%s' database. Error: '%v'; DSN: '%s'", driverName, err, connectionStringForLogs)
	}
}
