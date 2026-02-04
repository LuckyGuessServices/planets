package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/paths/general"
)

const (
	VarNameApplicationEnvironment string = "LUG_PLANETS_APP_ENV"
	defaultApplicationEnvironment        = ApplicationEnvironmentDevLocal

	VarNameApplicationFilesRootDirectory = "LUG_PLANETS_APP_FILES_DIR"

	// DBCore

	VarNameDBCoreDatabaseName     string = "LUG_PLANETS_DBCORE_DBNAME"
	defaultDBCoreDatabaseName     string = "planets"
	defaultDBCoreDatabaseNameTest string = "planets_test"

	VarNameDBCoreHost string = "LUG_PLANETS_DBCORE_HOST"
	defaultDBCoreHost string = "127.0.0.1"

	VarNameDBCorePort string = "LUG_PLANETS_DBCORE_PORT"
	defaultDBCorePort string = "5400"

	VarNameDBCoreUsername     string = "LUG_PLANETS_DBCORE_USERNAME"
	defaultDBCoreUsername     string = "planets_user"
	defaultDBCoreUsernameTest string = "planets_test_user"

	VarNameDBCorePassword     string = "LUG_PLANETS_DBCORE_PASSWORD"
	defaultDBCorePassword     string = "planets_pass"
	defaultDBCorePasswordTest string = "planets_test_pass"

	VarNameDBCoreRootDatabaseName string = "LUG_PLANETS_DBCORE_ROOT_DBNAME"
	defaultDBCoreRootDatabaseName string = "postgres"

	VarNameDBCoreRootUsername string = "LUG_PLANETS_DBCORE_ROOT_USERNAME"
	defaultDBCoreRootUsername string = "root_user"

	VarNameDBCoreRootPassword string = "LUG_PLANETS_DBCORE_ROOT_PASSWORD"
	defaultDBCoreRootPassword string = "root_pass"

	// SERVER

	VarNameServerListenHost string = "LUG_PLANETS_SERVER_LISTEN_HOST"
	defaultServerListenHost string = "127.0.0.1"

	VarNameServerListenPort string = "LUG_PLANETS_SERVER_LISTEN_PORT"
	defaultServerListenPort string = "8080"
)

// ConfigStruct defines fields for all environment variables needed for the application.
type ConfigStruct struct {
	applicationEnvironment        ApplicationEnvironmentType
	applicationFilesRootDirectory string

	dbCoreDatabaseName     string
	dbCoreHost             string
	dbCorePort             int
	dbCoreUsername         string
	dbCorePassword         string
	dbCoreRootDatabaseName string
	dbCoreRootUsername     string
	dbCoreRootPassword     string

	serverListenHost string
	serverListenPort int
}

func (config *ConfigStruct) ApplicationEnvironment() ApplicationEnvironmentType {
	return config.applicationEnvironment
}

func (config *ConfigStruct) ApplicationFilesRootDirectory() string {
	return config.applicationFilesRootDirectory
}

func (config *ConfigStruct) DBCoreDatabaseName() string {
	return config.dbCoreDatabaseName
}

func (config *ConfigStruct) DBCoreHost() string {
	return config.dbCoreHost
}

func (config *ConfigStruct) DBCorePort() int {
	return config.dbCorePort
}

func (config *ConfigStruct) DBCoreUsername() string {
	return config.dbCoreUsername
}

func (config *ConfigStruct) DBCorePassword() string {
	return config.dbCorePassword
}

func (config *ConfigStruct) DBCoreRootDatabaseName() string {
	return config.dbCoreRootDatabaseName
}

func (config *ConfigStruct) DBCoreRootUsername() string {
	return config.dbCoreRootUsername
}

func (config *ConfigStruct) DBCoreRootPassword() string {
	return config.dbCoreRootPassword
}

func (config *ConfigStruct) ServerListenHost() string {
	return config.serverListenHost
}

func (config *ConfigStruct) ServerListenPort() int {
	return config.serverListenPort
}

var config *ConfigStruct

// Config initializes the only [ConfigStruct] instance with values read from environment variables.
// If you want to alter some values before usage within the application, do not call this function
// until all necessary [os.Setenv] calls are made.
func Config() *ConfigStruct {
	if config != nil {
		return config
	}

	appEnvRaw := varValue(VarNameApplicationEnvironment, string(defaultApplicationEnvironment), false)
	appEnv := ApplicationEnvironmentType(appEnvRaw)
	if err := validateAppEnv(appEnv); err != nil {
		luglog.Panic(err)
	}

	appRootDir := varValue(VarNameApplicationFilesRootDirectory, "", true)
	if appRootDir == "" {
		_, thisFilePath, _, isOk := runtime.Caller(0)
		if !isOk {
			luglog.Panic("Failed to determine project source root directory.")
		}

		appRootDir = general.Abs(filepath.Dir(thisFilePath) + "/../..")
	}

	dbCorePortString := varValue(VarNameDBCorePort, defaultDBCorePort, false)
	dbCorePortInt, errDBCorePortConv := strconv.Atoi(dbCorePortString)
	if errDBCorePortConv != nil {
		luglog.Panicf(
			"Failed to convert CORE db port value '%s' to an integer from env var '%s': %v",
			dbCorePortString,
			VarNameDBCorePort,
			errDBCorePortConv,
		)
	}

	serverPortString := varValue(VarNameServerListenPort, defaultServerListenPort, false)
	serverPortInt, errServerPortConv := strconv.Atoi(serverPortString)
	if errServerPortConv != nil {
		luglog.Panicf(
			"Failed to convert server port value '%s' to an integer from env var '%s': %v",
			serverPortString,
			VarNameServerListenPort,
			errServerPortConv,
		)
	}

	var dbCoreDatabaseName, dbCoreUsername, dbCorePassword string
	if isTestInternal(appEnv) {
		dbCoreDatabaseName = varValue(VarNameDBCoreDatabaseName, defaultDBCoreDatabaseNameTest, false)
		dbCoreUsername = varValue(VarNameDBCoreUsername, defaultDBCoreUsernameTest, false)
		dbCorePassword = varValue(VarNameDBCorePassword, defaultDBCorePasswordTest, false)
	} else {
		dbCoreDatabaseName = varValue(VarNameDBCoreDatabaseName, defaultDBCoreDatabaseName, false)
		dbCoreUsername = varValue(VarNameDBCoreUsername, defaultDBCoreUsername, false)
		dbCorePassword = varValue(VarNameDBCorePassword, defaultDBCorePassword, false)
	}

	config = &ConfigStruct{
		applicationEnvironment:        appEnv,
		applicationFilesRootDirectory: appRootDir,

		dbCoreDatabaseName:     dbCoreDatabaseName,
		dbCoreHost:             varValue(VarNameDBCoreHost, defaultDBCoreHost, false),
		dbCorePort:             dbCorePortInt,
		dbCoreUsername:         dbCoreUsername,
		dbCorePassword:         dbCorePassword,
		dbCoreRootDatabaseName: varValue(VarNameDBCoreRootDatabaseName, defaultDBCoreRootDatabaseName, false),
		dbCoreRootUsername:     varValue(VarNameDBCoreRootUsername, defaultDBCoreRootUsername, false),
		dbCoreRootPassword:     varValue(VarNameDBCoreRootPassword, defaultDBCoreRootPassword, false),

		serverListenHost: varValue(VarNameServerListenHost, defaultServerListenHost, true),
		serverListenPort: serverPortInt,
	}

	return config
}

type ApplicationEnvironmentType string

const (
	ApplicationEnvironmentDevLocal ApplicationEnvironmentType = "app_dev_local"
	ApplicationEnvironmentTest     ApplicationEnvironmentType = "app_test"
	ApplicationEnvironmentProd     ApplicationEnvironmentType = "app_prod"
)

func validateAppEnv(appEnvValue ApplicationEnvironmentType) error {
	switch appEnvValue {
	case ApplicationEnvironmentDevLocal, ApplicationEnvironmentTest, ApplicationEnvironmentProd:
		return nil
	default:
		return fmt.Errorf("unsupported environment type ('%v') value: %v", VarNameApplicationEnvironment, appEnvValue)
	}
}

func IsTest() bool {
	return isTestInternal(Config().ApplicationEnvironment())
}

func isTestInternal(appEnv ApplicationEnvironmentType) bool {
	return appEnv == ApplicationEnvironmentTest
}

func PanicIfEnvIsTest() {
	if IsTest() {
		luglog.Panic(
			"Application environment must be any of non-test environments. Current: ",
			Config().ApplicationEnvironment(),
		)
	}
}

func PanicIfEnvNotTest() {
	if !IsTest() {
		luglog.Panic(
			"Application environment must be any of test environments. Current: ",
			Config().ApplicationEnvironment(),
		)
	}
}

func varValue(envVarName string, defaultValue string, allowEmpty bool) string {
	value, isPresent := os.LookupEnv(envVarName)
	if !isPresent {
		return defaultValue
	}

	value = strings.TrimSpace(value)
	if !allowEmpty && value == "" {
		luglog.Panicf("'%v' env var must not be empty", envVarName)
	}

	return value
}
