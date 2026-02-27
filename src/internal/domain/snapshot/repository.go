package snapshot

import (
	"context"

	"github.com/LuckyGuessServices/planets/internal/domain/planets"
	"github.com/LuckyGuessServices/planets/internal/general/types"
	"github.com/LuckyGuessServices/planets/internal/integration/mercury"
	"github.com/LuckyGuessServices/planets/internal/luglog"
)

type RepositoryInterface interface {
	NewSnapshot() *Snapshot
	BySnapshotDate(ctx context.Context, snapshotDate types.Date) ([]*Snapshot, error)

	// ObtainSnapshot obtains (by calculating or requesting external sources) the necessary data
	// and adds a new snapshot to a repository.
	ObtainSnapshot(snapshotDate types.Date)
}

// RepositoryCommons represents a set of common methods shared with all implementations.
type RepositoryCommons struct {
	RepositoryInterface
}

func (repo *RepositoryCommons) ObtainSnapshot(snapshotDate types.Date) {
	// Here in the future we will create a background task (RabbitMQ / Kafka / etc.).
	ctxMercury := context.Background()

	snapshotMercury := repo.NewSnapshot()
	snapshotMercury.SetSnapshotDate(snapshotDate)
	snapshotMercury.SetPlanetIndex(planets.Mercury().Index())

	isRetrograde, errMercury := mercury.Client().ByDate(ctxMercury, snapshotDate)
	if errMercury == nil {
		snapshotMercury.SetIsRetrograde(isRetrograde)
	}

	if err := snapshotMercury.Save(ctxMercury); err != nil {
		luglog.Panicf(`Failed to store new Mercury snapshot by date "%s": %v`, snapshotDate, err)
	}
}
