package fastjson

import (
	"bytes"
	"testing"
)

func generateWhitespaceData(size int) []byte {
	return bytes.Repeat([]byte(" \t\n\r"), size/4)
}

func naiveSkipWhitespace(data []byte) int {
	i := 0
	for i < len(data) {
		c := data[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
		} else {
			break
		}
	}
	return i
}

func BenchmarkSkipWhiteSpace_LUT(b *testing.B) {
	data := generateWhitespaceData(1024 * 1024)

	b.ResetTimer()
	for b.Loop() {
		it := NewIterator(data)
		it.skipWhiteSpace()
		if it.head != len(data) {
			b.Fatalf("Expected nothing, got %x at offset %d", it.char(), it.head)
		}
	}
}

func BenchmarkSkipWhiteSpace_Naive(b *testing.B) {
	data := generateWhitespaceData(1024 * 1024)

	b.ResetTimer()
	for b.Loop() {
		it := NewIterator(data)
		naiveSkipWhitespace(data)
		if it.head != len(data) {
			b.Fatalf("Expected nothing, got %x at offset %d", it.char(), it.head)
		}
	}
}

func BenchmarkSkipWhiteSpace_StdLib_TrimLeft(b *testing.B) {
	data := generateWhitespaceData(1024 * 1024)
	cutset := " \t\n\r"
	b.ResetTimer()

	for b.Loop() {
		res := bytes.TrimLeft(data, cutset)
		if len(res) != 0 {
			b.Fatalf("Expected nothing, got %s", res)
		}
	}
}
