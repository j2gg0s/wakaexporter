package wakaexporter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/util/retry"
)

// SyncToPG
// NOTE: force is not real force
func SyncToPG(ctx context.Context, apiKey string, from, to time.Time, db *pg.DB, force bool) error {
	from, to = from.Truncate(24*time.Hour), to.Truncate(24*time.Hour)

	heartbeat := Heartbeat{}
	err := db.ModelContext(ctx, &heartbeat).Order("created_at desc").First()
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return fmt.Errorf("query pg with error: %w", err)
	} else if errors.Is(err, pg.ErrNoRows) {
		// pass
	} else if synced := heartbeat.CreatedAt.Truncate(24 * time.Hour); synced.After(from) && !force {
		from = synced
	}

	if from.After(to) {
		return nil
	}

	for date := from; !date.After(to); date = date.Add(24 * time.Hour) {
		var heartbeats []Heartbeat
		err := retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
			var err error
			heartbeats, err = ListHeartbeat(ctx, apiKey, date)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		rows, err := InsertHeartbeats(ctx, db, heartbeats)
		if err != nil {
			return err
		}

		log.Debug().Msgf("%v: get heartbeats %d, updated: %d", date, len(heartbeats), rows)
	}
	return nil
}
