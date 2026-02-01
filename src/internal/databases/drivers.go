package databases

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DATA-DOG/go-txdb"
	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const DBMainDriverName = "pgx"

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

func OpenAndPingPGX(ctx context.Context, dataSourceName string, debugInfo string) *pgxpool.Pool {
	db, errOpen := pgxpool.New(ctx, dataSourceName)
	if errOpen != nil {
		luglog.Panicf("Unable to open PGX Pool. Error: '%v'; DSN: '%s'", errOpen, debugInfo)
	}
	if err := db.Ping(ctx); err != nil {
		luglog.Panicf("Unable to ping PGX Pool. Error: '%v'; DSN: '%s'", err, debugInfo)
	}

	return db
}

func OpenAndPing(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	db, errOpen := sql.Open(driverName, dataSourceName)
	if errOpen != nil {
		luglog.Panicf("Unable to open '%s' database. Error: '%v'; DSN: '%s'", driverName, errOpen, debugInfo)
	}
	ping(db, driverName, debugInfo)

	return db
}

// OpenAndPingTxDB opens a modified database pool. This pool should be used in tests only.
//
// All queries made with this pool are enclosed in a single global transaction.
// This transaction is rolled back as soon as the pool is closed.
//
// Opening and commiting / roll backing transactions within this pool actually operates with safe points.
// The single global transactions stays intact and always rolls back.
func OpenAndPingTxDB(driverName string, dataSourceName string, debugInfo string) *sql.DB {
	env.PanicIfEnvNotTest()

	dbTxDB := sql.OpenDB(txdb.New(driverName, dataSourceName))
	ping(dbTxDB, driverName+"(TxDB)", debugInfo)

	return dbTxDB
}

func ping(db *sql.DB, driverName string, debugInfo string) {
	if err := db.Ping(); err != nil {
		_ = db.Close()
		luglog.Panicf("Unable to ping '%s' database. Error: '%v'; DSN: '%s'", driverName, err, debugInfo)
	}
}
