package fastjson

import (
	"fmt"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type Iterator struct {
	head    int
	data    []byte
	dataLen int
}

func NewIterator(data []byte) *Iterator {
	return &Iterator{
		head:    0,
		data:    data,
		dataLen: len(data),
	}
}

func (it *Iterator) Reset(data []byte) {
	it.head = 0
	it.data = data
	it.dataLen = len(data)
}

func (it *Iterator) error(msg string) error {
	return fmt.Errorf("fastjson: %s at offset %d", msg, it.head)
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (it *Iterator) skipWhiteSpace() {
	for it.head < it.dataLen {
		if parseTable[it.data[it.head]]&maskWhiteSpace == 0 {
			break
		}
		it.head++
	}
}

// --- Structural Helpers ---
func (it *Iterator) ReadObjectStart() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == '{' {
		it.head++
		return nil
	}
	return it.error("expected '{'")
}

func (it *Iterator) ReadObjectEnd() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == '}' {
		it.head++
		return nil
	}
	return it.error("expected '}'")
}

func (it *Iterator) ReadArrayStart() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == '[' {
		it.head++
		return nil
	}
	return it.error("expected '['")
}

func (it *Iterator) ReadArrayEnd() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == ']' {
		it.head++
		return nil
	}
	return it.error("expected ']'")
}

func (it *Iterator) ReadComma() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == ',' {
		it.head++
		return nil
	}
	return it.error("expected ','")
}

func (it *Iterator) ReadColon() error {
	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == ':' {
		it.head++
		return nil
	}
	return it.error("expected ':'")
}

// --- Primitive Parsers ---
func (it *Iterator) ReadInt64() (int64, error) {
	it.skipWhiteSpace()
	if it.head >= it.dataLen {
		return 0, it.error("unexpected end of input")
	}

	neg := false
	switch it.data[it.head] {
	case '-':
		neg = true
		it.head++
	case '+':
		return 0, it.error("leading '+' is not allowed in JSON numbers")
	}

	start := it.head
	var n int64 = 0

	if it.head < it.dataLen && it.data[it.head] == '0' {
		it.head++

		if it.head < it.dataLen {
			c := it.data[it.head]
			if c >= '0' && c <= '9' {
				return 0, it.error("leading zero is not allowed in JSON Numbers")
			}
		}
	} else {
		for it.head < it.dataLen {
			c := it.data[it.head]
			if c >= '0' && c <= '9' {
				n = n*10 + int64(c-'0')
				it.head++
			} else {
				break
			}
		}
	}

	if it.head == start {
		return 0, it.error("expected digit")
	}

	if it.head < it.dataLen {
		c := it.data[it.head]
		if c == '.' || c == 'e' || c == 'E' {
			return 0, it.error("float found, expected integer")
		}
	}

	if neg {
		n = -n
	}

	return n, nil
}

func (it *Iterator) ReadFloat64() (float64, error) {
	it.skipWhiteSpace()
	start := it.head

	for it.head < it.dataLen {
		c := it.data[it.head]
		if parseTable[c]&maskNumber == 0 {
			break
		}
		it.head++
	}

	if it.head == start {
		return 0, it.error("expected digit")
	}

	numStr := bytesToString(it.data[start:it.head])
	return strconv.ParseFloat(numStr, 64)
}

func (it *Iterator) ReadBool() (bool, error) {
	it.skipWhiteSpace()
	if it.head >= it.dataLen {
		return false, it.error("unexpected end of input")
	}

	if it.data[it.head] == 't' {
		if it.head+4 <= it.dataLen && bytesToString(it.data[it.head:it.head+4]) == "true" {
			it.head += 4
			return true, nil
		}
		return false, it.error("expected 'true'")
	}

	if it.data[it.head] == 'f' {
		if it.head+5 <= it.dataLen && bytesToString(it.data[it.head:it.head+5]) == "false" {
			it.head += 5
			return false, nil
		}
		return false, it.error("expected 'false'")
	}

	return false, it.error("expected boolean")
}

func (it *Iterator) ReadNull() error {
	it.skipWhiteSpace()
	if it.head >= it.dataLen {
		return it.error("unexpected end of input")
	}

	if it.data[it.head] == 'n' {
		if it.head+3 <= it.dataLen && bytesToString(it.data[it.head:it.head+3]) == "null" {
			it.head += 4
			return nil
		}
		return it.error("expected 'null'")
	}

	return it.error("expected null")
}

func (it *Iterator) ReadString() (string, error) {
	it.skipWhiteSpace()
	if it.head >= it.dataLen {
		return "", it.error("unexpected end of input")
	}
	if it.data[it.head] != '"' {
		return "", it.error("expected start of string '\"'")
	}

	it.head++
	start := it.head
	for it.head < it.dataLen {
		c := it.data[it.head]
		if stringTable[c] != 0 {
			if c == '"' {
				str := bytesToString(it.data[start:it.head])
				it.head++
				return str, nil
			}
			if c == '\\' {
				return it.readStringSlow(start)
			}
			if c < 0x20 {
				return "", it.error("control character in string")
			}
		}
		it.head++
	}
	return "", it.error("unexpected end of input")
}

func (it *Iterator) readStringSlow(start int) (string, error) {
	out := make([]byte, 0, (len(it.data)-start)+16)
	out = append(out, it.data[start:it.head]...)
	for it.head < len(it.data) {
		c := it.data[it.head]
		if c == '"' {
			it.head++
			return string(out), nil
		}
		if c == '\\' {
			it.head++
			if it.head >= len(it.data) {
				return "", it.error("unexpected end of input in escape")
			}
			escape := it.data[it.head]
			switch escape {
			case '"', '\\', '/':
				out = append(out, escape)
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'u':
				if it.head+4 >= len(it.data) {
					return "", it.error("incomplete unicode escape")
				}
				r, err := it.decodeUnicode()
				if err != nil {
					return "", err
				}
				var buf [4]byte
				n := utf8.EncodeRune(buf[:], r)
				out = append(out, buf[:n]...)
				it.head += 4
			default:
				return "", it.error("invalid escape sequence")
			}
			it.head++
			continue
		}
		if c < 0x20 {
			return "", it.error("control character in string")
		}
		out = append(out, c)
		it.head++
	}
	return "", it.error("unexpected end of input in string")
}

func (it *Iterator) decodeUnicode() (rune, error) {
	start := it.head + 1
	if start+4 > len(it.data) {
		return 0, it.error("incomplete unicode escape")
	}
	var r rune
	for i := range 4 {
		c := it.data[start+i]
		var val rune
		if c >= '0' && c <= '9' {
			val = rune(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val = rune(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			val = rune(c - 'A' + 10)
		} else {
			return 0, it.error("invalid unicode hex digit")
		}
		r = (r << 4) | val
	}
	return r, nil
}

func (it *Iterator) char() byte {
	if it.head >= it.dataLen {
		return 0
	}

	return it.data[it.head]
}
