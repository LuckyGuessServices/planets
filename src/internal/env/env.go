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

	// DBMain

	VarNameDBMainDatabaseName     string = "LUG_PLANETS_DBMAIN_DBNAME"
	defaultDBMainDatabaseName     string = "planets"
	defaultDBMainDatabaseNameTest string = "planets_test"

	VarNameDBMainHost string = "LUG_PLANETS_DBMAIN_HOST"
	defaultDBMainHost string = "127.0.0.1"

	VarNameDBMainPort string = "LUG_PLANETS_DBMAIN_PORT"
	defaultDBMainPort string = "5400"

	VarNameDBMainUsername     string = "LUG_PLANETS_DBMAIN_USERNAME"
	defaultDBMainUsername     string = "planets_user"
	defaultDBMainUsernameTest string = "planets_test_user"

	VarNameDBMainPassword     string = "LUG_PLANETS_DBMAIN_PASSWORD"
	defaultDBMainPassword     string = "planets_pass"
	defaultDBMainPasswordTest string = "planets_test_pass"

	VarNameDBMainRootDatabaseName string = "LUG_PLANETS_DBMAIN_ROOT_DBNAME"
	defaultDBMainRootDatabaseName string = "postgres"

	VarNameDBMainRootUsername string = "LUG_PLANETS_DBMAIN_ROOT_USERNAME"
	defaultDBMainRootUsername string = "root_user"

	VarNameDBMainRootPassword string = "LUG_PLANETS_DBMAIN_ROOT_PASSWORD"
	defaultDBMainRootPassword string = "root_pass"

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

	dbMainDatabaseName     string
	dbMainHost             string
	dbMainPort             int
	dbMainUsername         string
	dbMainPassword         string
	dbMainRootDatabaseName string
	dbMainRootUsername     string
	dbMainRootPassword     string

	serverListenHost string
	serverListenPort int
}

func (config *ConfigStruct) ApplicationEnvironment() ApplicationEnvironmentType {
	return config.applicationEnvironment
}

func (config *ConfigStruct) ApplicationFilesRootDirectory() string {
	return config.applicationFilesRootDirectory
}

func (config *ConfigStruct) DBMainDatabaseName() string {
	return config.dbMainDatabaseName
}

func (config *ConfigStruct) DBMainHost() string {
	return config.dbMainHost
}

func (config *ConfigStruct) DBMainPort() int {
	return config.dbMainPort
}

func (config *ConfigStruct) DBMainUsername() string {
	return config.dbMainUsername
}

func (config *ConfigStruct) DBMainPassword() string {
	return config.dbMainPassword
}

func (config *ConfigStruct) DBMainRootDatabaseName() string {
	return config.dbMainRootDatabaseName
}

func (config *ConfigStruct) DBMainRootUsername() string {
	return config.dbMainRootUsername
}

func (config *ConfigStruct) DBMainRootPassword() string {
	return config.dbMainRootPassword
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
	if nil == config {
		appEnvRaw := varValue(VarNameApplicationEnvironment, string(defaultApplicationEnvironment), false)
		appEnv := ApplicationEnvironmentType(appEnvRaw)
		if err := validateAppEnv(appEnv); err != nil {
			luglog.Panic(err)
		}

		appRootDir := varValue(VarNameApplicationFilesRootDirectory, "", true)
		if "" == appRootDir {
			_, thisFilePath, _, isOk := runtime.Caller(0)
			if !isOk {
				luglog.Panic("Failed to determine project source root directory.")
			}

			appRootDir = general.Abs(filepath.Dir(thisFilePath) + "/../..")
		}

		dbMainPortString := varValue(VarNameDBMainPort, defaultDBMainPort, false)
		dbMainPortInt, errDBMainPortConv := strconv.Atoi(dbMainPortString)
		if errDBMainPortConv != nil {
			luglog.Panicf(
				"Failed to convert DBMain port value '%s' to an integer from env var '%s': %v",
				dbMainPortString,
				VarNameDBMainPort,
				errDBMainPortConv,
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

		var dbMainDatabaseName, dbMainUsername, dbMainPassword string
		if ApplicationEnvironmentTest == appEnv {
			dbMainDatabaseName = varValue(VarNameDBMainDatabaseName, defaultDBMainDatabaseNameTest, false)
			dbMainUsername = varValue(VarNameDBMainUsername, defaultDBMainUsernameTest, false)
			dbMainPassword = varValue(VarNameDBMainPassword, defaultDBMainPasswordTest, false)
		} else {
			dbMainDatabaseName = varValue(VarNameDBMainDatabaseName, defaultDBMainDatabaseName, false)
			dbMainUsername = varValue(VarNameDBMainUsername, defaultDBMainUsername, false)
			dbMainPassword = varValue(VarNameDBMainPassword, defaultDBMainPassword, false)
		}

		config = &ConfigStruct{
			applicationEnvironment:        appEnv,
			applicationFilesRootDirectory: appRootDir,

			dbMainDatabaseName:     dbMainDatabaseName,
			dbMainHost:             varValue(VarNameDBMainHost, defaultDBMainHost, false),
			dbMainPort:             dbMainPortInt,
			dbMainUsername:         dbMainUsername,
			dbMainPassword:         dbMainPassword,
			dbMainRootDatabaseName: varValue(VarNameDBMainRootDatabaseName, defaultDBMainRootDatabaseName, false),
			dbMainRootUsername:     varValue(VarNameDBMainRootUsername, defaultDBMainRootUsername, false),
			dbMainRootPassword:     varValue(VarNameDBMainRootPassword, defaultDBMainRootPassword, false),

			serverListenHost: varValue(VarNameServerListenHost, defaultServerListenHost, true),
			serverListenPort: serverPortInt,
		}
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

func PanicIfEnvNotTest() {
	appEnv := Config().ApplicationEnvironment()
	if ApplicationEnvironmentTest != appEnv {
		luglog.Panicf("Invalid application environment: '%v'. Expected: '%v' ", appEnv, ApplicationEnvironmentTest)
	}
}

func varValue(envVarName string, defaultValue string, allowEmpty bool) string {
	value, isPresent := os.LookupEnv(envVarName)
	if !isPresent {
		return defaultValue
	}

	value = strings.TrimSpace(value)
	if !allowEmpty && "" == value {
		luglog.Panicf("'%v' env var must not be empty", envVarName)
	}

	return value
}
