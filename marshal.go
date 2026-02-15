package fastjson

import (
	"reflect"
	"unsafe"
)

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	w := GetWriter()
	defer PutWriter(w)

	// Reflection is unavoidable at the very top level to unwrap the interface{}
	rv := reflect.ValueOf(v)
	t := rv.Type()

	var ptr unsafe.Pointer
	var enc EncoderFunc
	var err error

	if t.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		enc, err = getEncoder(t.Elem())
		ptr = unsafe.Pointer(rv.Pointer())
	} else {
		enc, err = getEncoder(t)

		newPtr := reflect.New(t)
		newPtr.Elem().Set(rv)
		ptr = unsafe.Pointer(newPtr.Pointer())
	}

	if err != nil {
		return nil, err
	}

	if err := enc(w, ptr); err != nil {
		return nil, err
	}

	// Copy the result buffer to return ownership to caller
	// (Since we reuse the Writer in a pool, we can't return w.Buffer directly)
	result := make([]byte, len(w.Buffer))
	copy(result, w.Buffer)
	return result, nil
}
