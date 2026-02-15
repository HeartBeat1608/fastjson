package fastjson

import (
	"testing"
)

func TestReadString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"Simple", `"hello"`, "hello", false},
		{"Empty", `""`, "", false},
		{"With Spaces", `"hello world"`, "hello world", false},
		{"Escaped Quote", `"hello \"world\""`, `hello "world"`, false},
		{"Escaped Newline", `"line1\nline2"`, "line1\nline2", false},
		{"Escaped Backslash", `"C:\\path"`, `C:\path`, false},
		{"Unicode Basic", `"\u0041"`, "A", false},
		{"Unicode Emoji", `"\u263A"`, "☺", false}, // Smiley face
		{"Mixed", `"Val: \u0041!"`, "Val: A!", false},
		{"Missing Quote", `"hello`, "", true},
		{"Control Char", "\"\n\"", "", true}, // Real newline inside string is invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewIterator([]byte(tt.input))
			val, err := it.ReadString()

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if val != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, val)
			}
		})
	}
}
