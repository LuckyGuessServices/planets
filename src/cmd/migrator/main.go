package main

import (
	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/shutdown_cleanup"
)

func main() {
	defer shutdown_cleanup.ExecuteStack()

	databases.MigrateUp()
	// databases.MigrateDownTo(0)
	// databases.GenerateMigrationScript("Create table snapshots")
}
