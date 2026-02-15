package fastjson

import (
	"encoding/json"
	"testing"
)

func BenchmarkReadInt64_Fast(b *testing.B) {
	data := []byte("1234567890")
	b.ResetTimer()
	for b.Loop() {
		it := NewIterator(data)
		_, err := it.ReadInt64()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadInt64_StdLib(b *testing.B) {
	data := []byte("1234567890")
	var v int64
	b.ResetTimer()
	for b.Loop() {
		// encoding/json usually unmarshals numbers to float64 in interface{},
		// but here we force it to int64 field
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadFloat64_Fast(b *testing.B) {
	data := []byte("12345.6789")
	b.ResetTimer()
	for b.Loop() {
		it := NewIterator(data)
		_, err := it.ReadFloat64()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadFloat64_StdLib(b *testing.B) {
	data := []byte("12345.6789")
	var v float64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}
