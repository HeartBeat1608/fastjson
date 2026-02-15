package fastjson

import (
	"reflect"
	"unsafe"
)

func Unmarshal(data []byte, v any) error {
	it := NewIterator(data)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return it.error("Unmarshal(non-pointer or nil)")
	}

	dec, err := getDecoder(rv.Elem().Type())
	if err != nil {
		return err
	}

	ptr := unsafe.Pointer(rv.Pointer())
	return dec(it, ptr)
}
