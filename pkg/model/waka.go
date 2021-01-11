package model

import "time"

type Heartbeat struct {
	ID string `json:"id,omitempty" pg:",pk"`

	UserID   string `json:"user_id,omitempty"`
	Project  string `json:"project,omitempty"`
	Language string `json:"language,omitempty"`

	Category string `json:"category,omitempty"`
	Entity   string `json:"entity,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	Branch        string   `json:"branch,omitempty"`
	CursorPOS     int      `json:"cursorpos,omitempty" pg:"cursorpos"`
	Dependencies  []string `json:"dependencies,omitempty" pg:",array"`
	IsWrite       bool     `json:"is_write"`
	LineNO        int      `json:"lineno,omitempty" pg:"lineno"`
	Lines         int      `json:"lines,omitempty"`
	MachineNameID string   `json:"machine_name_id,omitempty"`
	Time          float64  `json:"time,omitempty"`
	Type          string   `json:"type,omitempty"`
	UserAgentID   string   `json:"user_agent_id,omitempty"`

	tableName struct{} `pg:"heartbeat"` //nolint
}

type Metric struct {
	Time   time.Time              `pg:",pk"`
	Name   string                 `pg:",pk"`
	Labels map[string]interface{} `pg:",pk"`
	Value  float64

	tableName struct{} `pg:"metric"` // nolint
}
