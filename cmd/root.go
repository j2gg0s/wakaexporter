package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/go-pg/pg/v10"
)

var (
	debug  bool = false
	apiKey string

	pgDSN string
	pgDB  *pg.DB
)

func NewRootCommand() *cobra.Command {
	root := cobra.Command{
		Use: "wakaexporter",
	}

	root.PersistentFlags().BoolVar(&debug, "debug", debug, "enable debug log")

	root.PersistentFlags().StringVar(&apiKey, "api-key", apiKey, "wakatime's secret api key")

	root.PersistentFlags().StringVar(&pgDSN, "pg", pgDSN, "dsn of postgresql, if you want to save heartbeats to postgresql")

	root.PersistentPreRunE = func(*cobra.Command, []string) error {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		if len(apiKey) > 0 {
			apiKey = base64.StdEncoding.EncodeToString([]byte(apiKey))
		}

		opts, err := pg.ParseURL(pgDSN)
		if err != nil {
			return fmt.Errorf("invalid pg dsn %s: %w", pgDSN, err)
		}
		pgDB = pg.Connect(opts)

		return nil
	}

	return &root
}
