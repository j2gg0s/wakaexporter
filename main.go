package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-pg/pg/v10"

	service "github.com/j2gg0s/wakaexporter/pkg/service/wakaexporter"
)

var (
	debug    bool = false
	apiKey   string
	fromDate string = time.Now().Format("2006-01-02")
	toDate   string = fromDate
	from, to time.Time

	force bool = false

	pgDSN    string
	pgClient *pg.DB
)

func main() {
	root := cobra.Command{
		Use: "wakaexporter",
	}

	root.PersistentFlags().BoolVar(&debug, "debug", debug, "enable debug log")

	root.PersistentFlags().StringVar(&apiKey, "api-key", apiKey, "wakatime's secret api key")
	root.PersistentFlags().StringVar(&fromDate, "from-date", fromDate, "scrape heartbeats from, include")
	root.PersistentFlags().StringVar(&toDate, "to-date", toDate, "scrape heartbeats to, include")

	root.PersistentFlags().StringVar(&pgDSN, "pg", pgDSN, "dsn of postgresql, if you want to save heartbeats to postgresql")

	root.PersistentFlags().BoolVar(&force, "force", force, "force refresh heartbeats")

	root.PersistentPreRunE = func(*cobra.Command, []string) error {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		if len(apiKey) == 0 {
			return fmt.Errorf("api-key is required")
		}
		apiKey = base64.StdEncoding.EncodeToString([]byte(apiKey))

		var err error
		from, err = time.Parse("2006-01-02", fromDate)
		if err != nil {
			return fmt.Errorf("fromDate %s is not valid: %w", fromDate, err)
		}
		to, err = time.Parse("2006-01-02", toDate)
		if err != nil {
			return fmt.Errorf("toDate %s is not valid: %s", toDate, err)
		}
		if from.After(to) {
			return fmt.Errorf("invalid time range: %s -> %s", fromDate, toDate)
		}

		return nil
	}

	root.RunE = func(*cobra.Command, []string) error {
		opts, err := pg.ParseURL(pgDSN)
		if err != nil {
			return fmt.Errorf("invalid pg dsn %s: %w", pgDSN, err)
		}
		pgClient = pg.Connect(opts)

		return service.SyncToPG(context.Background(), apiKey, from, to, pgClient, force)
	}

	if err := root.Execute(); err != nil {
		log.Err(err).Send()
	}
}
