package wakaexporter

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"k8s.io/client-go/util/retry"

	"github.com/go-pg/pg/v10"
	"github.com/j2gg0s/wakaexporter/pkg/client"
	"github.com/j2gg0s/wakaexporter/pkg/model"
	"github.com/j2gg0s/wakaexporter/pkg/orm"
	"github.com/rs/zerolog/log"
)

// don't sync data within 10m
var minDuration = 10 * time.Minute

func SyncHeartbeat(ctx context.Context, db *pg.DB, apiKey string) error {
	hb, err := orm.GetLastHeartbeat(ctx, db)
	if err != nil {
		log.Err(err).Send()
		// default sync one month's data
		hb = &model.Heartbeat{CreatedAt: time.Now().Add(-30 * 24 * time.Hour)}
	}

	begin := hb.CreatedAt
	if time.Since(begin) < minDuration {
		log.Debug().Msgf("sync data succeed at %s", begin.Format(time.RFC3339))
		return nil
	}

	for t := time.Now(); t.After(begin); t = t.Add(-24 * time.Hour) {
		var hbs []model.Heartbeat
		err := retry.OnError(retry.DefaultBackoff, func(error) bool { return true }, func() error {
			var err error
			hbs, err = client.ListHeartbeat(ctx, apiKey, t)
			if err != nil {
				log.Warn().Err(err).Msgf("get heartbeat failed and retry")
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("get heartbeat failed and failed: %w", err)
		}
		row := 0
		if len(hbs) > 0 {
			row, err = orm.InsertHeartbeats(ctx, db, hbs)
			if err != nil {
				return err
			}
		}
		log.Debug().Msgf("insert %d heartbeats at %s", row, t.Format("2006-01-02"))
	}

	return nil
}

func StatsdHeartbeats(ctx context.Context, hbs []model.Heartbeat) ([]model.Metric, error) {
	metrics := []model.Metric{}

	var x, y model.Heartbeat
	var xProj, yProj string
	x, xProj = hbs[0], getProject(hbs[0], "")
	for i := 1; i < len(hbs); i += 1 {
		y, yProj = hbs[i], getProject(hbs[i], xProj)
		// wakatime send heartbeat every 120 second
		if xProj == yProj && y.Time-x.Time > 120 {
			if (y.Time - x.Time) > 120*1.25 {
				log.Warn().Msgf("duration is too bigger: %f", y.Time-x.Time)
			}
			x, xProj = y, yProj
			continue
		}
		value := y.Time - x.Time
		if math.Abs(value) < 0.01 {
			// ignore too small
			continue
		}

		sec := int64(x.Time)
		nsec := int64((x.Time - float64(sec)) * float64(time.Second/time.Nanosecond))
		metrics = append(metrics, model.Metric{
			Time: time.Unix(sec, nsec),
			Name: "heartbeat",
			Labels: map[string]interface{}{
				"user":     x.UserID,
				"category": x.Category,
				"project":  xProj,
			},
			Value: y.Time - x.Time,
		})

		x, xProj = y, yProj
	}

	return metrics, nil
}

func getProject(h model.Heartbeat, curr string) string {
	p := h.Project
	if h.Category == "browsing" {
		p = h.Entity
	} else if p == "" && (strings.Contains(h.Entity, "/go/pkg") || strings.Contains(h.Entity, "/go/src")) {
		p = curr
	}
	return p
}

func SyncMetric(ctx context.Context, db *pg.DB) error {
	m, err := orm.GetLastMetric(ctx, db)
	if err != nil {
		return err
	}

	hbs, err := orm.ListHeartbeat(ctx, db, m.Time.Add(-1*minDuration), time.Now().Add(-1*minDuration))
	if err != nil {
		return err
	}

	metrics, err := StatsdHeartbeats(ctx, hbs)
	if err != nil {
		return nil
	}

	rows := 0
	if len(metrics) > 0 {
		rows, err = orm.InsertMetrics(ctx, db, metrics)
		if err != nil {
			return nil
		}
	}
	log.Debug().Msgf("insert %d metric with %d heartbeats", rows, len(hbs))

	return nil
}
