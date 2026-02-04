package databases_test

import (
	"context"
	"testing"

	"github.com/LuckyGuessServices/planets/internal/databases"
	"github.com/LuckyGuessServices/planets/internal/test_tools"
	"github.com/uptrace/bun"
)

func TestMain(m *testing.M) {
	test_tools.RunTestMain(m)
}

// Tests no prepared statements are made via both master and readonly database pools.
func TestPGXDoesNotPrepareStatementsInsideDbPool(t *testing.T) {
	tests := []struct {
		name         string
		isDBReadonly bool
	}{
		{name: "CORE", isDBReadonly: false},
		{name: "CORE readonly", isDBReadonly: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test_tools.RunInSyncBubble(t, func(t *testing.T) {
				var dbPool *bun.DB
				if test.isDBReadonly {
					dbPool = databases.CoreReadonly()
				} else {
					dbPool = databases.Core()
				}

				ctx := context.Background()

				var notUsed int
				if err := dbPool.NewSelect().ColumnExpr("? + ?", 1, 1).Scan(ctx, &notUsed); err != nil {
					t.Fatalf("Failed to make a query: %v", err)
				}

				var count int
				errSelectPrepared := dbPool.NewSelect().
					TableExpr("pg_prepared_statements").
					ColumnExpr("count(*)").
					Scan(ctx, &count)
				if errSelectPrepared != nil {
					t.Fatalf("Failed to query pg_prepared_statements data: %v", errSelectPrepared)
				}
				if count > 0 {
					t.Fatalf("Expected 0 prepared statements, found %d", count)
				}
			})
		})
	}
}
