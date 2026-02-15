package fastjson

import (
	"encoding/json"
	"testing"
)

type SmallStruct struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	IsActive bool    `json:"is_active"`
	Score    float64 `json:"score"`
}

type LargeStruct struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Items       []SmallStruct `json:"items"`
	Meta        struct {
		Count int    `json:"count"`
		Type  string `json:"type"`
	} `json:"meta"`
	Tags []string `json:"tags"`
}

// Generate sample data
func getSmallJSON() []byte {
	return []byte(`{"id": 12345, "name": "Benchmark User", "is_active": true, "score": 99.9}`)
}

func getLargeJSON() []byte {
	// A reasonably complex object with nested arrays and structs
	return []byte(`{
		"title": "Performance Test",
		"description": "Benchmarking JSON parsers in Go",
		"items": [
			{"id": 1, "name": "Item 1", "is_active": true, "score": 10.5},
			{"id": 2, "name": "Item 2", "is_active": false, "score": 20.5},
			{"id": 3, "name": "Item 3", "is_active": true, "score": 30.5},
			{"id": 4, "name": "Item 4", "is_active": true, "score": 40.5},
			{"id": 5, "name": "Item 5", "is_active": false, "score": 50.5}
		],
		"meta": {
			"count": 5,
			"type": "dataset"
		},
		"tags": ["golang", "json", "performance", "benchmark", "fast"]
	}`)
}

// --- Benchmarks ---

// 1. Small Struct Decoding
func BenchmarkUnmarshal_SmallStruct_Fast(b *testing.B) {
	data := getSmallJSON()
	var s SmallStruct
	b.ResetTimer()
	for b.Loop() {
		if err := Unmarshal(data, &s); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_SmallStruct_Std(b *testing.B) {
	data := getSmallJSON()
	var s SmallStruct
	b.ResetTimer()
	for b.Loop() {
		if err := json.Unmarshal(data, &s); err != nil {
			b.Fatal(err)
		}
	}
}

// 2. Large Complex Struct Decoding
func BenchmarkUnmarshal_LargeStruct_Fast(b *testing.B) {
	data := getLargeJSON()
	var l LargeStruct
	b.ResetTimer()
	for b.Loop() {
		if err := Unmarshal(data, &l); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_LargeStruct_Std(b *testing.B) {
	data := getLargeJSON()
	var l LargeStruct
	b.ResetTimer()
	for b.Loop() {
		if err := json.Unmarshal(data, &l); err != nil {
			b.Fatal(err)
		}
	}
}

// 3. Map[string]any (Generic) Decoding
func BenchmarkUnmarshal_Interface_Fast(b *testing.B) {
	data := getSmallJSON()
	var v any
	b.ResetTimer()
	for b.Loop() {
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Interface_Std(b *testing.B) {
	data := getSmallJSON()
	var v any
	b.ResetTimer()
	for b.Loop() {
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// 4. Map[string]any (Generic) Decoding - Large
func BenchmarkUnmarshal_InterfaceLarge_Fast(b *testing.B) {
	data := getLargeJSON()
	var v any
	b.ResetTimer()
	for b.Loop() {
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_InterfaceLarge_Std(b *testing.B) {
	data := getLargeJSON()
	var v any
	b.ResetTimer()
	for b.Loop() {
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}
