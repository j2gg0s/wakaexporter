package wakaexporter

import "time"

type Heartbeat struct {
	ID string `json:"id,omitempty" pg:",pk"`

	UserID   string `json:"user_id,omitempty"`
	Project  string `json:"project,omitempty"`
	Language string `json:"language,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`

	Branch        string   `json:"branch,omitempty"`
	Category      string   `json:"category,omitempty"`
	CursorPOS     int      `json:"cursorpos,omitempty" pg:"cursorpos"`
	Dependencies  []string `json:"dependencies,omitempty" pg:",array"`
	Entity        string   `json:"entity,omitempty"`
	IsWrite       bool     `json:"is_write"`
	LineNO        int      `json:"lineno,omitempty" pg:"lineno"`
	Lines         int      `json:"lines,omitempty"`
	MachineNameID string   `json:"machine_name_id,omitempty"`
	Time          float64  `json:"time,omitempty"`
	Type          string   `json:"type,omitempty"`
	UserAgentID   string   `json:"user_agent_id,omitempty"`

	tableName struct{} `pg:"heartbeat"` //nolint
}
