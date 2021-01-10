package wakaexporter

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
)

func InsertHeartbeats(ctx context.Context, db *pg.DB, heartbeats []Heartbeat) (int, error) {
	res, err := db.ModelContext(ctx, &heartbeats).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return 0, fmt.Errorf("bulk insert heartbeats with error: %w", err)
	}
	return res.RowsAffected(), nil
}
