package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	service "github.com/j2gg0s/wakaexporter/pkg/service/wakaexporter"
)

func NewSyncCommand() *cobra.Command {
	cmd := cobra.Command{
		Use: "sync",
	}

	var (
		spec string = "0 * * * *"
	)

	cmd.PersistentFlags().StringVar(&spec, "cron", spec, "support standard cron spec: https://en.wikipedia.org/wiki/Cron")

	cmd.PreRunE = func(*cobra.Command, []string) error {
		if pgDB == nil {
			return fmt.Errorf("pg is required: %s", pgDSN)
		}

		return nil
	}

	cmd.RunE = func(*cobra.Command, []string) error {
		c := cron.New()
		_, err := c.AddFunc(spec, func() {
			ctx := context.Background()
			now := time.Now()

			if err := service.SyncHeartbeat(ctx, pgDB, apiKey); err != nil {
				log.Warn().Err(err).Msgf("sync heartbeats failed at %s", now)
				return
			}
			if err := service.SyncMetric(ctx, pgDB); err != nil {
				log.Warn().Err(err).Msgf("sync metrics failed at %s", now)
				return
			}

			log.Debug().Msgf("sync heartbeat && metric succeed at %s", now)
		})
		if err != nil {
			return fmt.Errorf("add %s to cron with error: %w", spec, err)
		}

		c.Run()

		return nil
	}

	return &cmd
}
