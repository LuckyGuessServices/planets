package databases

import (
	"github.com/LuckyGuessServices/planets/internal/luglog"
	"github.com/LuckyGuessServices/planets/internal/paths"
	_ "github.com/LuckyGuessServices/planets/migrations"
	"github.com/pressly/goose/v3"
)

func MigrateUp() {
	if err := goose.Up(Core().DB, paths.MigrationScriptsDir(), goose.WithAllowMissing()); err != nil {
		luglog.Panic("Failed to migrate up to the last migration: ", err)
	}
}

//goland:noinspection GoUnusedExportedFunction
func MigrateDownTo(version int64) {
	if err := goose.DownTo(Core().DB, paths.MigrationScriptsDir(), version); err != nil {
		luglog.Panicf("Failed to migrate down to '%d': %v", version, err)
	}
}

//goland:noinspection GoUnusedExportedFunction
func GenerateMigrationScript(migrationName string) {
	if err := goose.Create(Core().DB, paths.MigrationScriptsDir(), migrationName, "go"); err != nil {
		luglog.Panic("Failed to generate a migration script: ", err)
	}
}
