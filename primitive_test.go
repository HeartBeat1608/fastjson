package fastjson

import (
	"math"
	"testing"
)

func TestReadInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		fail     bool
	}{
		{"Simple", "123", 123, false},
		{"Negative", "-123", -123, false},
		{"Zero", "0", 0, false},
		{"Large", "9223372036854775807", 9223372036854775807, false},
		{"With Space", "  42 ", 42, false},
		{"Leading Zero", "01", 0, true},   // Invalid JSON
		{"Positive Sign", "+1", 0, true},  // Invalid JSON
		{"Float as Int", "12.3", 0, true}, // Should fail Int64
		{"Empty", "", 0, true},
		{"Chars", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewIterator([]byte(tt.input))
			val, err := it.ReadInt64()
			if tt.fail {
				if err == nil {
					t.Errorf("expected error, got %d", val)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if val != tt.expected {
					t.Errorf("expected %d, got %d", tt.expected, val)
				}
			}
		})
	}
}

func TestReadFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Simple", "123.45", 123.45},
		{"Negative", "-123.45", -123.45},
		{"Scientific", "1.23e2", 123.0},
		{"Big E", "1E-1", 0.1},
		{"Integer", "100", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewIterator([]byte(tt.input))
			val, err := it.ReadFloat64()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(val-tt.expected) > 0.000001 {
				t.Errorf("expected %f, got %f", tt.expected, val)
			}
		})
	}
}

func TestReadBool(t *testing.T) {
	it := NewIterator([]byte("true false"))

	v1, err := it.ReadBool()
	if err != nil || !v1 {
		t.Errorf("expected true")
	}

	v2, err := it.ReadBool()
	if err != nil || v2 {
		t.Errorf("expected false")
	}
}
