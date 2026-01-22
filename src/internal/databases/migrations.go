package databases

import (
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/paths"
	"github.com/pressly/goose/v3"
)

func MigrateUp() {
	err := goose.Up(
		Main(),
		paths.MigrationScriptsDir(),
		goose.WithAllowMissing(),
	)
	if err != nil {
		luglog.Panic("Failed to migrate up to the last migration: ", err)
	}
}

func MigrateDownOne() {
	err := goose.Down(
		Main(),
		paths.MigrationScriptsDir(),
	)
	if err != nil {
		luglog.Panic("Failed to migrate down a single migration: ", err)
	}
}
