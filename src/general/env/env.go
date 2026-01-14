package env

import (
	"os"
	"strings"

	"github.com/LuckyGuessServices/planets/general/luglog"
)

const (
	VarNameApplicationEnvironment string = "LUG_APP_ENV"

	VarNameServerListenHost string = "LUG_SERVER_LISTEN_HOST"
	VarNameServerListenPort string = "LUG_SERVER_LISTEN_PORT"

	VarNameDBMainDatabaseName string = "LUG_DBMAIN_DBNAME"
	VarNameDBMainHost         string = "LUG_DBMAIN_HOST"
	VarNameDBMainPort         string = "LUG_DBMAIN_PORT"
	VarNameDBMainUsername     string = "LUG_DBMAIN_USERNAME"
	VarNameDBMainPassword     string = "LUG_DBMAIN_PASSWORD"
)

func VarValue(envVarName string, defaultValue string, allowEmpty bool) string {
	value, isPresent := os.LookupEnv(envVarName)
	if !isPresent {
		return defaultValue
	}

	value = strings.TrimSpace(value)
	if !allowEmpty && "" == value {
		luglog.Fatalf("'%v' env var must not be empty", envVarName)
	}

	return value
}
