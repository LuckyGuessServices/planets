package main

import (
	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
)

func main() {
	defer shutdown_cleanup.ExecuteStack()

	databases.MigrateUp()
	// databases.MigrateDownOne()
}
