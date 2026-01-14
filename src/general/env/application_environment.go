package env

import (
	"LuG/planets/general/luglog"
	"fmt"
)

type ApplicationEnvironmentType string

const (
	ApplicationEnvironmentDevLocal ApplicationEnvironmentType = "app_dev_local"
	ApplicationEnvironmentTest     ApplicationEnvironmentType = "app_test"
	ApplicationEnvironmentProd     ApplicationEnvironmentType = "app_prod"
)

func ValidateAppEnv(appEnvValue ApplicationEnvironmentType) error {
	switch appEnvValue {
	case ApplicationEnvironmentDevLocal, ApplicationEnvironmentTest, ApplicationEnvironmentProd:
		return nil
	default:
		return fmt.Errorf("unsupported environment type ('%v') value: %v", VarNameApplicationEnvironment, appEnvValue)
	}
}

var appEnv ApplicationEnvironmentType

func AppEnv() ApplicationEnvironmentType {
	if "" == appEnv {
		initAppEnv()
	}

	return appEnv
}

func initAppEnv() {
	envTypeString := VarValue(VarNameApplicationEnvironment, string(ApplicationEnvironmentDevLocal), false)
	appEnv = ApplicationEnvironmentType(envTypeString)
	if err := ValidateAppEnv(appEnv); err != nil {
		luglog.Fatal(err)
	}
}
