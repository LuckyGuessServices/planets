package snapshot

import (
	"context"

	"github.com/LuckyGuessServices/planets/internal/general/types"
)

type RepositoryInterface interface {
	NewSnapshot() *Snapshot
	BySnapshotDate(ctx context.Context, snapshotDate types.Date) ([]*Snapshot, error)
}
