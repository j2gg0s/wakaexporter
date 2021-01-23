package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	service "github.com/j2gg0s/wakaexporter/pkg/service/wakaexporter"
)

func NewRefreshCommand() *cobra.Command {
	cmd := cobra.Command{
		Use: "refresh",
	}

	var (
		fromDate string = time.Now().Format("2006-01-02")
		toDate   string = fromDate
		from, to time.Time
	)
	cmd.PersistentFlags().StringVar(&fromDate, "from-date", fromDate, "scrape heartbeats from, include")
	cmd.PersistentFlags().StringVar(&toDate, "to-date", toDate, "scrape heartbeats to, include")

	cmd.PreRunE = func(*cobra.Command, []string) error {
		if pgDB == nil {
			return fmt.Errorf("pg is required: %s", pgDSN)
		}

		var err error
		from, err = time.Parse("2006-01-02", fromDate)
		if err != nil {
			return fmt.Errorf("fromDate %s is not valid: %w", fromDate, err)
		}
		from = from.Truncate(24 * time.Hour)

		to, err = time.Parse("2006-01-02", toDate)
		if err != nil {
			return fmt.Errorf("toDate %s is not valid: %s", toDate, err)
		}
		to = to.Truncate(24 * time.Hour).Add(time.Hour * 24)

		if from.After(to) {
			return fmt.Errorf("invalid time range: %s -> %s", fromDate, toDate)
		}

		return nil
	}

	cmd.RunE = func(*cobra.Command, []string) error {
		return service.RefreshMetric(context.Background(), pgDB, from, to)
	}

	return &cmd
}
