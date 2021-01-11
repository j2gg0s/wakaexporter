package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/j2gg0s/wakaexporter/pkg/model"
)

func InsertHeartbeats(ctx context.Context, db *pg.DB, heartbeats []model.Heartbeat) (int, error) {
	res, err := db.ModelContext(ctx, &heartbeats).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return 0, fmt.Errorf("bulk insert heartbeats with error: %w", err)
	}
	return res.RowsAffected(), nil
}

func InsertMetrics(ctx context.Context, db *pg.DB, metrics []model.Metric) (int, error) {
	res, err := db.ModelContext(ctx, &metrics).Insert()
	if err != nil {
		return 0, fmt.Errorf("bulk insert metrics with error: %w", err)
	}
	return res.RowsAffected(), nil
}

func GetLastHeartbeat(ctx context.Context, db *pg.DB) (*model.Heartbeat, error) {
	hb := model.Heartbeat{}
	if err := db.ModelContext(ctx, &hb).Order("created_at desc").First(); err != nil {
		return nil, fmt.Errorf("query heartbeat from pg with error: %w", err)
	}
	return &hb, nil
}

func GetLastMetric(ctx context.Context, db *pg.DB) (*model.Metric, error) {
	metric := model.Metric{}
	err := db.ModelContext(ctx, &metric).Order("time desc").First()
	if err != nil {
		return nil, fmt.Errorf("query metric from pg with error: %w", err)
	}
	return &metric, nil
}

func ListHeartbeat(ctx context.Context, db *pg.DB, from, to time.Time) ([]model.Heartbeat, error) {
	hbs := []model.Heartbeat{}

	err := db.ModelContext(ctx, &hbs).
		Where("time >= ?", from.Unix()).
		Order("time asc").
		Select()
	if err != nil {
		return nil, fmt.Errorf("query heartbeats from pg with error: %w", err)
	}

	return hbs, nil
}
