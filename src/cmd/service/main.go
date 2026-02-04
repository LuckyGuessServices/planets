package main

import (
	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/lugserver"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
)

func main() {
	env.PanicIfEnvIsTest()

	defer shutdown_cleanup.ExecuteStack()

	lugserver.Serve()
}
