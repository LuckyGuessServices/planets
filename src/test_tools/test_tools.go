package test_tools

import (
	"os"
	"testing"

	"github.com/LuckyGuessServices/planets/general/env"
	"github.com/LuckyGuessServices/planets/general/luglog"
	"github.com/LuckyGuessServices/planets/general/shutdown_cleanup"
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

	if err := os.Setenv(env.VarNameApplicationEnvironment, string(env.ApplicationEnvironmentTest)); err != nil {
		luglog.Fatal(err)
	}
	if env.ApplicationEnvironmentTest != env.AppEnv() {
		luglog.Fatal("Invalid environment type:", env.AppEnv())
	}
}
