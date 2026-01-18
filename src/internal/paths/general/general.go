package general

import (
	"path/filepath"

	"github.com/LuckyGuessServices/planets/internal/luglog"
)

// Abs returns an absolute path to a file or directory.
//
// Panics, if an error occurs.
func Abs(somePath string) string {
	absolutePath, err := filepath.Abs(somePath)
	if err != nil {
		luglog.Fatalf("Failed to detect absolute path for: '%s'", somePath)
	}

	return absolutePath
}
