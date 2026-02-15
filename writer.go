package fastjson

import (
	"strconv"
	"sync"
)

type Writer struct {
	Buffer []byte
}

var hexChars = "0123456789abcdef"

var writerPool = sync.Pool{
	New: func() any {
		return &Writer{Buffer: make([]byte, 0, 512)}
	},
}

func GetWriter() *Writer {
	w := writerPool.Get().(*Writer)
	w.Buffer = w.Buffer[:0]
	return w
}

func PutWriter(w *Writer) {
	writerPool.Put(w)
}

func (w *Writer) Write(p []byte) {
	w.Buffer = append(w.Buffer, p...)
}

func (w *Writer) WriteByte(c byte) {
	w.Buffer = append(w.Buffer, c)
}

func (w *Writer) WriteString(s string) {
	w.Buffer = append(w.Buffer, s...)
}

// WriteStringEscaped writes a string with JSON quoting and escaping.
// It uses the Look-up Table (stringTable) to optimize the 'happy path'.
func (w *Writer) WriteStringEscaped(s string) {
	w.Buffer = append(w.Buffer, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		if stringTable[s[i]] != 0 {
			if i > start {
				w.Buffer = append(w.Buffer, s[start:i]...)
			}
			w.writeEscapedChar(s[i])
			start = i + 1
		}
	}

	if start < len(s) {
		w.Buffer = append(w.Buffer, s[start:]...)
	}

	w.Buffer = append(w.Buffer, '"')
}

func (w *Writer) writeEscapedChar(c byte) {
	switch c {
	case '"':
		w.Buffer = append(w.Buffer, '\\', '"')
	case '\\':
		w.Buffer = append(w.Buffer, '\\', '\\')
	case '\b':
		w.Buffer = append(w.Buffer, '\\', 'b')
	case '\f':
		w.Buffer = append(w.Buffer, '\\', 'f')
	case '\n':
		w.Buffer = append(w.Buffer, '\\', 'n')
	case '\r':
		w.Buffer = append(w.Buffer, '\\', 'r')
	case '\t':
		w.Buffer = append(w.Buffer, '\\', 't')
	default:
		if c < 0x20 {
			w.Buffer = append(w.Buffer, '\\', 'u', '0', '0')
			w.Buffer = append(w.Buffer, hexChars[c>>4], hexChars[c&0xf])
		} else {
			w.Buffer = append(w.Buffer, c)
		}
	}
}

func (w *Writer) WriteInt64(n int64) {
	w.Buffer = strconv.AppendInt(w.Buffer, n, 10)
}

func (w *Writer) WriteUint64(n uint64) {
	w.Buffer = strconv.AppendUint(w.Buffer, n, 10)
}

func (w *Writer) WriteFloat64(n float64) {
	b := w.Buffer
	b = strconv.AppendFloat(b, n, 'f', -1, 64)
	w.Buffer = b
}

func (w *Writer) WriteBool(b bool) {
	if b {
		w.Buffer = append(w.Buffer, "true"...)
	} else {
		w.Buffer = append(w.Buffer, "false"...)
	}
}

func (w *Writer) WriteNull() {
	w.Buffer = append(w.Buffer, "null"...)
}
