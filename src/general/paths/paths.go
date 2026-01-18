package paths

import (
	"fmt"

	"github.com/LuckyGuessServices/planets/general/env"
)

func MigrationScriptsDir() string {
	return fmt.Sprintf("%s/%s", env.Config().ApplicationFilesRootDirectory(), "migrations")
}
