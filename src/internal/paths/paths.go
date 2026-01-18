package paths

import (
	"fmt"

	"github.com/LuckyGuessServices/planets/internal/env"
)

func MigrationScriptsDir() string {
	return fmt.Sprintf("%s/%s", env.Config().ApplicationFilesRootDirectory(), "migrations")
}
