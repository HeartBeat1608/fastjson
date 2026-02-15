package fastjson

import "testing"

func TestSkipWhiteSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected byte
	}{
		{"Simple Space", "   {", '{'},
		{"Tabs and Newlines", "\t\n\r [", '['},
		{"No Whitespace", "1234", '1'},
		{"Mixed", "  \n\t  \"", '"'},
		{"Empty", "", 0},
		{"Only Whitespae", "     ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewIterator([]byte(tt.input))
			it.skipWhiteSpace()
			if it.char() != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, it.head)
			}
		})
	}
}
