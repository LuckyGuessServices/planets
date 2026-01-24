package paths

import (
	"github.com/LuckyGuessServices/planets/internal/env"
)

func MigrationScriptsDir() string {
	return env.Config().ApplicationFilesRootDirectory() + "/migrations"
}
