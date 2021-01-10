package wakaexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const addr = "https://wakatime.com/api/v1/users"

func ListHeartbeat(ctx context.Context, apiKey string, date time.Time) ([]Heartbeat, error) {
	params := url.Values{}
	params.Set("date", date.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(
		ctx, "GET", fmt.Sprintf("%s/current/heartbeats?%s", addr, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("new request wtih error: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s with error: %w", req.URL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request %s without 200: %d", req.URL.String(), resp.StatusCode)
	}

	body := struct {
		Heartbeats []Heartbeat `json:"data"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("unmarshal json with error: %w", err)
	}

	return body.Heartbeats, nil
}
