package fastjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

// --- Production Data Model ---

type APIResponse struct {
	Status    int        `json:"status"`
	Message   string     `json:"message"`
	ServerID  string     `json:"server_id"`
	Timestamp int64      `json:"timestamp"`
	Data      []APIUser  `json:"data"`
	Meta      Pagination `json:"meta"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type APIUser struct {
	ID          int               `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	IsActive    bool              `json:"is_active"`
	Roles       []string          `json:"roles"`
	Preferences map[string]string `json:"preferences"`
	LastLogin   string            `json:"last_login"` // ISO8601 string
	RiskScore   float64           `json:"risk_score"`
}

// --- Data Generators ---

func generatePayload(itemCount int) *APIResponse {
	users := make([]APIUser, itemCount)
	for i := range itemCount {
		users[i] = APIUser{
			ID:          1000 + i,
			Username:    fmt.Sprintf("user_%d", i),
			Email:       fmt.Sprintf("user_%d@example.com", i),
			IsActive:    i%2 == 0,
			Roles:       []string{"viewer", "editor", "admin"},                                 // 3-item slice
			Preferences: map[string]string{"theme": "dark", "lang": "en-US", "notifs": "push"}, // 3-item map
			LastLogin:   "2023-10-27T10:00:00Z",
			RiskScore:   0.12345 + float64(i),
		}
	}

	return &APIResponse{
		Status:    200,
		Message:   "Request successful",
		ServerID:  "srv-aws-useast-1a-992",
		Timestamp: 1698400000,
		Data:      users,
		Meta: Pagination{
			Page:       1,
			PerPage:    itemCount,
			TotalItems: itemCount * 10,
			TotalPages: 10,
		},
	}
}

// Pre-compute JSON bytes for Unmarshal benchmarks
var (
	// smallPayloadObj  = generatePayload(10)   // ~3KB
	mediumPayloadObj = generatePayload(100)  // ~30KB (Typical API Page)
	largePayloadObj  = generatePayload(1000) // ~300KB

	// smallPayloadBytes, _  = json.Marshal(smallPayloadObj)
	mediumPayloadBytes, _ = json.Marshal(mediumPayloadObj)
	largePayloadBytes, _  = json.Marshal(largePayloadObj)
)

// --- Unmarshal Benchmarks (Reading) ---

func BenchmarkAPI_Unmarshal_Medium_FastJSON(b *testing.B) {
	var resp APIResponse
	b.SetBytes(int64(len(mediumPayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if err := Unmarshal(mediumPayloadBytes, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Unmarshal_Medium_StdLib(b *testing.B) {
	var resp APIResponse
	b.SetBytes(int64(len(mediumPayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if err := json.Unmarshal(mediumPayloadBytes, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Unmarshal_Large_FastJSON(b *testing.B) {
	var resp APIResponse
	b.SetBytes(int64(len(largePayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if err := Unmarshal(largePayloadBytes, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Unmarshal_Large_StdLib(b *testing.B) {
	var resp APIResponse
	b.SetBytes(int64(len(largePayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if err := json.Unmarshal(largePayloadBytes, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Marshal Benchmarks (Writing) ---

func BenchmarkAPI_Marshal_Medium_FastJSON(b *testing.B) {
	b.SetBytes(int64(len(mediumPayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if _, err := Marshal(mediumPayloadObj); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Marshal_Medium_StdLib(b *testing.B) {
	b.SetBytes(int64(len(mediumPayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if _, err := json.Marshal(mediumPayloadObj); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Marshal_Large_FastJSON(b *testing.B) {
	b.SetBytes(int64(len(largePayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if _, err := Marshal(largePayloadObj); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPI_Marshal_Large_StdLib(b *testing.B) {
	b.SetBytes(int64(len(largePayloadBytes)))
	b.ResetTimer()

	for b.Loop() {
		if _, err := json.Marshal(largePayloadObj); err != nil {
			b.Fatal(err)
		}
	}
}
