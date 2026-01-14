package main

import (
	"LuG/planets/general/lugserver"
	"LuG/planets/general/shutdown_cleanup"
)

func main() {
	defer shutdown_cleanup.ExecuteStack()

	lugserver.Serve()
}
