package storage

import (
	"github.com/LuckyGuessServices/planets/internal/domain/snapshot"
)

var repository snapshot.RepositoryInterface

func SnapshotRepository() snapshot.RepositoryInterface {
	if repository != nil {
		return repository
	}

	repository = (&SnapshotsBunRepository{}).Init()

	return repository
}
