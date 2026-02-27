package storage

import (
	"context"

	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/uptrace/bun"
)

type SnapshotsBunRepository struct {
	snapshot.RepositoryCommons
}

func (repo *SnapshotsBunRepository) Init() *SnapshotsBunRepository {
	repo.RepositoryInterface = repo

	return repo
}

func (repo *SnapshotsBunRepository) NewSnapshot() *snapshot.Snapshot {
	return &snapshot.Snapshot{EntityInterface: &SnapshotBun{}}
}

func (repo *SnapshotsBunRepository) BySnapshotDate(
	ctx context.Context,
	snapshotDate types.Date,
) ([]*snapshot.Snapshot, error) {
	rowset := make([]*SnapshotBun, 0)
	errSelect := databases.CoreReadonly().NewSelect().
		Model(&rowset).
		Where("?TableAlias.snapshot_date = ?", snapshotDate).
		OrderExpr("?TableAlias.planet_index", bun.OrderAsc).
		Scan(ctx)
	if errSelect != nil {
		return nil, errSelect
	}

	collection := make([]*snapshot.Snapshot, 0, len(rowset))
	for _, row := range rowset {
		collection = append(collection, &snapshot.Snapshot{EntityInterface: row})
	}

	return collection, nil
}
