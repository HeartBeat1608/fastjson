package fastjson

import (
	"encoding/json"
	"testing"
)

// Benchmark the fast path (no escapes)
func BenchmarkReadString_Simple(b *testing.B) {
	data := []byte(`"this_is_a_common_json_key_or_value"`)
	b.ResetTimer()

	for b.Loop() {
		it := NewIterator(data)
		_, err := it.ReadString()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark the slow path (escapes)
func BenchmarkReadString_Escaped(b *testing.B) {
	data := []byte(`"this\nis\ta\ntext\twith\nescapes"`)
	b.ResetTimer()

	for b.Loop() {
		it := NewIterator(data)
		_, err := it.ReadString()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Standard Library for comparison
func BenchmarkStdLib_UnmarshalString(b *testing.B) {
	data := []byte(`"this_is_a_common_json_key_or_value"`)
	var s string
	b.ResetTimer()

	for b.Loop() {
		if err := json.Unmarshal(data, &s); err != nil {
			b.Fatal(err)
		}
	}
}
