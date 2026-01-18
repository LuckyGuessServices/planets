package main

import (
	"github.com/LuckyGuessServices/planets/general/lugserver"
	"github.com/LuckyGuessServices/planets/general/shutdown_cleanup"
)

func main() {
	defer shutdown_cleanup.ExecuteStack()

	lugserver.Serve()
}
