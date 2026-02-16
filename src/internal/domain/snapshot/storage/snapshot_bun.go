package storage

import (
	"context"
	"time"

	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/uptrace/bun"
)

type SnapshotBun struct {
	bun.BaseModel `bun:"table:snapshots,alias:sna"`

	ColumnID           int                  `bun:"id,pk,autoincrement"`
	ColumnSnapshotDate types.Date           `bun:"snapshot_date,notnull"`
	ColumnPlanetIndex  int16                `bun:"planet_index,notnull"`
	ColumnIsRetrograde types.Nullable[bool] `bun:"is_retrograde"`
	ColumnCreatedAt    time.Time            `bun:"created_at,notnull"`
}

func (row *SnapshotBun) BeforeAppendModel(_ context.Context, query bun.Query) error {
	currentMoment := time.Now().UTC()

	switch query.(type) {
	case *bun.InsertQuery:
		if row.ColumnCreatedAt.IsZero() {
			row.ColumnCreatedAt = currentMoment
		}
	}

	return nil
}

func (row *SnapshotBun) ID() int {
	return row.ColumnID
}

func (row *SnapshotBun) PlanetIndex() int16 {
	return row.ColumnPlanetIndex
}

func (row *SnapshotBun) IsRetrogradeRaw() types.Nullable[bool] {
	return row.ColumnIsRetrograde
}

func (row *SnapshotBun) SetSnapshotDate(value types.Date) {
	row.ColumnSnapshotDate = value
}

func (row *SnapshotBun) SetPlanetIndex(value int16) {
	row.ColumnPlanetIndex = value
}

func (row *SnapshotBun) SetIsRetrograde(value bool) {
	row.ColumnIsRetrograde.V = value
	row.ColumnIsRetrograde.Valid = true
}

func (row *SnapshotBun) Save(ctx context.Context) (err error) {
	db := databases.Core()

	if row.ColumnID > 0 {
		_, err = db.NewUpdate().
			Model(row).
			WherePK().
			Exec(ctx)

		return
	}

	_, err = db.NewInsert().
		Model(row).
		Exec(ctx)

	return
}
