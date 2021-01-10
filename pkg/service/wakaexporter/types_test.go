package wakaexporter

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHaeartbeat(t *testing.T) {
	fixtures := []map[string]interface{}{
		{
			"category":        "browsing",
			"created_at":      "2021-01-08T16:00:53Z",
			"entity":          "https://www.iqiyi.com",
			"id":              "dc5e5791-12c7-4b20-aa53-00df99808af1",
			"is_write":        false,
			"machine_name_id": "27ddd948-d239-48cb-a62d-322c4596d28e",
			"project":         "kubenotify",
			"time":            1610121652.000000,
			"type":            "domain",
			"user_agent_id":   "b50b895b-70d6-4bf5-b0de-a51725898c0a",
			"user_id":         "6383f3a2-09da-4637-a7ab-440e6e86312f",
		},
		{
			"branch":     "master",
			"category":   "coding",
			"created_at": "2021-01-09T08:08:06Z",
			"dependencies": []string{
				"\"strconv\"",
				"\"k8s.io/apimachinery/pkg/apis/meta/v1\"",
				"\"k8s.io/client-go/tools/cache\"",
			},
			"entity":          "/Users/j2gg0s/go/src/github.com/j2gg0s/kubenotify/pkg/handler2/resoruce.go",
			"id":              "de2e83e9-d5ea-46da-97ad-186b8f001231",
			"is_write":        true,
			"language":        "Go",
			"lines":           75,
			"machine_name_id": "897bf4a4-91bd-411b-9b89-fff9134cf7f4",
			"project":         "kubenotify",
			"time":            1610179656.635632,
			"type":            "file",
			"user_agent_id":   "95d09acd-4f21-4078-8392-09abd8b985cc",
			"user_id":         "6383f3a2-09da-4637-a7ab-440e6e86312f",
		},
	}

	for i, f := range fixtures {
		fixture := f
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			b, err := json.Marshal(fixture)
			require.NoError(t, err)

			heartbeat := Heartbeat{}
			err = json.Unmarshal(b, &heartbeat)
			require.NoError(t, err)

			actualB, err := json.Marshal(heartbeat)
			require.NoError(t, err)
			require.Equal(t, b, actualB)
		})
	}
}
