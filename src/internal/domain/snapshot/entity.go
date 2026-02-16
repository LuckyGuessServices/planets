package snapshot

import (
	"context"

	"github.com/LuckyGuessServices/planets/internal/general/types"
)

type EntityInterface interface {
	ID() int
	PlanetIndex() int16
	// IsRetrogradeRaw returns the flag state that can also be null (unknown).
	// "IsRetrograde" method name is reserved for answering the exact question, when null is equivalent to false.
	IsRetrogradeRaw() types.Nullable[bool]

	SetSnapshotDate(value types.Date)
	SetPlanetIndex(value int16)
	SetIsRetrograde(value bool)

	Save(ctx context.Context) (err error)
}
