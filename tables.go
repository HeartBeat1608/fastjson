package fastjson

const (
	maskWhiteSpace = 1 << iota
	maskNumber
	maskString
	maskObjectStart
	maskObjectEnd
	maskArrayStart
	maskArrayEnd
	maskSeparator
)

var (
	parseTable  [256]byte
	stringTable [256]byte
)

func init() {
	for i := range 256 {
		c := byte(i)

		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			parseTable[i] |= maskWhiteSpace
		}

		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
			parseTable[i] |= maskNumber
		}
		if c == '"' {
			parseTable[i] |= maskString
		}

		switch c {
		case '{':
			parseTable[i] |= maskObjectStart
		case '}':
			parseTable[i] |= maskObjectEnd
		case '[':
			parseTable[i] |= maskArrayStart
		case ']':
			parseTable[i] |= maskArrayEnd
		case ':', ',':
			parseTable[i] |= maskSeparator
		}

		if c == '"' || c == '\\' || c < 0x20 {
			stringTable[i] = 1
		}
	}
}
