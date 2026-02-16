package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateTableSnapshots, downCreateTableSnapshots)
}

func upCreateTableSnapshots(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(
		ctx,
		`
			CREATE TABLE snapshots (
				id SERIAL PRIMARY KEY,
				snapshot_date DATE NOT NULL,
				planet_index SMALLINT NOT NULL,
				is_retrograde BOOLEAN NULL DEFAULT NULL,
				created_at TIMESTAMP NOT NULL,
				CONSTRAINT snapshot_date_planet_index_u_idx UNIQUE (snapshot_date, planet_index)
			);

			COMMENT ON COLUMN snapshots.snapshot_date IS 'Calendar date the snapshot data is relevant to.';
			COMMENT ON COLUMN snapshots.planet_index IS
			        'Planet index (starting from 1 as the closest to a star) the snapshot data is relevant to.';
			COMMENT ON COLUMN snapshots.is_retrograde IS
			        'If a planet is in retrograde state. "NULL" means the state is unknown.';
			COMMENT ON COLUMN snapshots.created_at IS 'Date and time when the record was added to the table.';
		`,
	)

	return err
}

func downCreateTableSnapshots(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS "snapshots"`)

	return err
}
