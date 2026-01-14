package env

import (
	"LuG/planets/general/luglog"
	"os"
	"strings"
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
