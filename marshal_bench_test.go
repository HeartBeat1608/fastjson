package fastjson

import (
	"encoding/json"
	"testing"
)

// Using same structs as previous benchmarks

func BenchmarkMarshal_SmallStruct_Fast(b *testing.B) {
	s := SmallStruct{ID: 12345, Name: "Benchmark User", IsActive: true, Score: 99.9}
	b.ResetTimer()
	for b.Loop() {
		_, err := Marshal(&s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_SmallStruct_Std(b *testing.B) {
	s := SmallStruct{ID: 12345, Name: "Benchmark User", IsActive: true, Score: 99.9}
	b.ResetTimer()
	for b.Loop() {
		_, err := json.Marshal(&s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_LargeStruct_Fast(b *testing.B) {
	l := LargeStruct{
		Title:       "Performance Test",
		Description: "Benchmarking JSON parsers in Go",
		Items: []SmallStruct{
			{ID: 1, Name: "Item 1", IsActive: true, Score: 10.5},
			{ID: 2, Name: "Item 2", IsActive: false, Score: 20.5},
		},
		Tags: []string{"golang", "json", "fast"},
	}
	b.ResetTimer()
	for b.Loop() {
		_, err := Marshal(&l)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_LargeStruct_Std(b *testing.B) {
	l := LargeStruct{
		Title:       "Performance Test",
		Description: "Benchmarking JSON parsers in Go",
		Items: []SmallStruct{
			{ID: 1, Name: "Item 1", IsActive: true, Score: 10.5},
			{ID: 2, Name: "Item 2", IsActive: false, Score: 20.5},
		},
		Tags: []string{"golang", "json", "fast"},
	}
	b.ResetTimer()
	for b.Loop() {
		_, err := json.Marshal(&l)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Slice_Fast(b *testing.B) {
	list := make([]SmallStruct, 100)
	for i := range 100 {
		list[i] = SmallStruct{ID: i, Name: "Item", IsActive: true, Score: float64(i)}
	}
	b.ResetTimer()
	for b.Loop() {
		_, err := Marshal(&list)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Slice_Std(b *testing.B) {
	list := make([]SmallStruct, 100)
	for i := range 100 {
		list[i] = SmallStruct{ID: i, Name: "Item", IsActive: true, Score: float64(i)}
	}
	b.ResetTimer()
	for b.Loop() {
		_, err := json.Marshal(&list)
		if err != nil {
			b.Fatal(err)
		}
	}
}
