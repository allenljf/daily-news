package contract

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCheckedInOpenAPIKeepsBearerSecurityOnEveryV1Operation(t *testing.T) {
	contents, err := os.ReadFile("../../docs/contracts/daily-news.openapi.json")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var document struct {
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	for path, operations := range document.Paths {
		if len(path) < 3 || path[:3] != "/v1" {
			continue
		}
		for method, operation := range operations {
			if len(operation.Security) != 1 {
				t.Fatalf("%s %s security = %#v, want HTTPBearer", method, path, operation.Security)
			}
			if _, ok := operation.Security[0]["HTTPBearer"]; !ok {
				t.Fatalf("%s %s security = %#v, want HTTPBearer", method, path, operation.Security)
			}
		}
	}
}
