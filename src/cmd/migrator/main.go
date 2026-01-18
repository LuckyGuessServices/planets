package main

import (
	"github.com/LuckyGuessServices/planets/general/databases"
	"github.com/LuckyGuessServices/planets/general/shutdown_cleanup"
)

func main() {
	defer shutdown_cleanup.ExecuteStack()

	databases.MigrateUp()
	//databases.MigrateDownOne()
}
