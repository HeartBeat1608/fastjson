package fastjson

import (
	"encoding/json"
	"testing"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Balance  float64
}

type ComplexUser struct {
	User
	Tags   []string `json:"tags"`
	Scores []int    `json:"scores"`
}

type MapUser struct {
	Properties map[string]string `json:"properties"`
	Stats      map[string]int    `json:"stats"`
}

func TestUnmarshal_Struct(t *testing.T) {
	jsonStr := `{"id": 101, "name": "Gemini", "is_active": true, "Balance": 99.99}`
	var u User
	if err := Unmarshal([]byte(jsonStr), &u); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if u.ID != 101 {
		t.Errorf("ID mismatch")
	}
}

func TestUnmarshal_Slice(t *testing.T) {
	jsonStr := `{
		"id": 1, 
		"tags": ["admin", "editor", "viewer"],
		"scores": [10, 20, 30]
	}`
	var u ComplexUser
	if err := Unmarshal([]byte(jsonStr), &u); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(u.Tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(u.Tags))
	}
	if u.Tags[0] != "admin" {
		t.Errorf("Expected 'admin', got %s", u.Tags[0])
	}
	if len(u.Scores) != 3 {
		t.Fatalf("Expected 3 scores, got %d", len(u.Scores))
	}
	if u.Scores[2] != 30 {
		t.Errorf("Expected 30, got %d", u.Scores[2])
	}
}

func TestUnmarshal_Map(t *testing.T) {
	jsonStr := `{
		"properties": {"color": "blue", "size": "large"},
		"stats": {"wins": 10, "losses": 2}
	}`
	var u MapUser
	if err := Unmarshal([]byte(jsonStr), &u); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if u.Properties["color"] != "blue" {
		t.Errorf("Expected blue, got %s", u.Properties["color"])
	}
	if u.Stats["wins"] != 10 {
		t.Errorf("Expected 10 wins, got %d", u.Stats["wins"])
	}
}

func BenchmarkUnmarshal_Slice(b *testing.B) {
	jsonStr := []byte(`{"tags": ["one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"]}`)
	b.Run("FastJSON", func(b *testing.B) {
		var u ComplexUser
		for b.Loop() {
			if err := Unmarshal(jsonStr, &u); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("StdLib", func(b *testing.B) {
		var u ComplexUser
		for b.Loop() {
			if err := json.Unmarshal(jsonStr, &u); err != nil {
				b.Fatal(err)
			}
		}
	})
}
