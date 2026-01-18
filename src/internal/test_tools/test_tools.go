package test_tools

import (
	"fmt"
	"os"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
)

var isGlobalSetUpLaunched = false

func RunTestMain(m *testing.M) {
	testsRunExitCode := 9999
	defer func() {
		shutdown_cleanup.ExecuteStack()

		os.Exit(testsRunExitCode)
	}()

	globalSetUp()

	testsRunExitCode = m.Run()
}

func globalSetUp() {
	if isGlobalSetUpLaunched {
		return
	}
	isGlobalSetUpLaunched = true

	enforceEnvVars()

	appEnv := env.Config().ApplicationEnvironment()
	if env.ApplicationEnvironmentTest != appEnv {
		luglog.Fatal("Invalid environment type: ", appEnv)
	}

	recreateTestDatabase()
	databases.MigrateUp()
}

func enforceEnvVars() {
	setEnvVar(env.VarNameApplicationEnvironment, string(env.ApplicationEnvironmentTest))
}

func recreateTestDatabase() {
	// CONFIG AND VARS ->

	envConfig := env.Config()

	dbConfig := databases.NewDBMainConfig()
	dbConfig.Host = envConfig.DBMainHost()
	dbConfig.Port = envConfig.DBMainPort()
	dbConfig.DBName = envConfig.DBMainRootDatabaseName()
	dbConfig.UserName = envConfig.DBMainRootUsername()
	dbConfig.UserPassword = envConfig.DBMainRootPassword()

	dbMainRoot := databases.Open(databases.DBMainDriverName, dbConfig.DSN(), dbConfig.DebugInfo())
	defer func() {
		err := dbMainRoot.Close()
		if err != nil {
			luglog.Fatal("[DB] Failed to close Main 'root' pool: ", err)
		} else {
			luglog.Print("[DB] Main 'root' pool is closed.")
		}
	}()
	luglog.Print("[DB] Main 'root' pool is initialized: ", dbConfig.DebugInfo())

	// <- CONFIG AND VARS

	// DATABASE RECREATION:

	_, errTerminateConnections := dbMainRoot.Exec(
		fmt.Sprintf(
			`
				SELECT pg_terminate_backend(pg_stat_activity.pid)
				FROM pg_stat_activity
				WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()
			`,
			envConfig.DBMainDatabaseName(),
		),
	)
	if errTerminateConnections != nil {
		luglog.Fatalf(
			"Failed to terminate existing connections to database '%s': %v",
			envConfig.DBMainDatabaseName(),
			errTerminateConnections,
		)
	}

	_, errDropDatabase := dbMainRoot.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", envConfig.DBMainDatabaseName()))
	if errDropDatabase != nil {
		luglog.Fatalf("Failed to drop database '%s': %v", envConfig.DBMainDatabaseName(), errDropDatabase)
	}

	_, errCreateDatabase := dbMainRoot.Exec(
		fmt.Sprintf("CREATE DATABASE \"%s\" OWNER \"%s\"", envConfig.DBMainDatabaseName(), envConfig.DBMainUsername()),
	)
	if errCreateDatabase != nil {
		luglog.Fatalf(
			"Failed to create database '%s' owned by user '%s': %v",
			envConfig.DBMainDatabaseName(),
			envConfig.DBMainUsername(),
			errCreateDatabase,
		)
	}

	luglog.Printf("Database '%s' is recreated.", envConfig.DBMainDatabaseName())
}

func setEnvVar(envVarName string, newValue string) {
	if err := os.Setenv(envVarName, newValue); err != nil {
		luglog.Fatalf(
			"Failed to set env var '%s' to '%v': %v",
			envVarName,
			newValue,
			err,
		)
	}
}
