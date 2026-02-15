package fastjson

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type EncoderFunc func(w *Writer, p unsafe.Pointer) error

var encoderCache sync.Map

func getEncoder(t reflect.Type) (EncoderFunc, error) {
	if f, ok := encoderCache.Load(t); ok {
		return f.(EncoderFunc), nil
	}

	enc, err := compileEncoder(t)
	if err != nil {
		return nil, err
	}

	encoderCache.Store(t, enc)
	return enc, nil
}

func compileEncoder(t reflect.Type) (EncoderFunc, error) {
	switch t.Kind() {
	case reflect.String:
		return encodeString, nil
	case reflect.Int, reflect.Int64:
		return encodeInt64, nil
	case reflect.Int32:
		return encodeInt32, nil
	case reflect.Float64:
		return encodeFloat64, nil
	case reflect.Bool:
		return encodeBool, nil
	case reflect.Interface:
		return encodeInterface, nil
	case reflect.Struct:
		return compileStructEncoder(t)
	case reflect.Slice:
		return compileSliceEncoderEnc(t)
	case reflect.Map:
		return compileMapEncoder(t)
	case reflect.Pointer:
		elemEnc, err := compileEncoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(w *Writer, p unsafe.Pointer) error {
			ptrVal := *(*unsafe.Pointer)(p)
			if ptrVal == nil {
				w.WriteNull()
				return nil
			}
			return elemEnc(w, ptrVal)
		}, nil
	default:
		return nil, fmt.Errorf("fastjson: unsupported type: %s", t.Kind())
	}
}

// Primitive encoders
func encodeString(w *Writer, p unsafe.Pointer) error {
	w.WriteStringEscaped(*(*string)(p))
	return nil
}

func encodeInt64(w *Writer, p unsafe.Pointer) error {
	w.WriteInt64(*(*int64)(p))
	return nil
}

func encodeInt32(w *Writer, p unsafe.Pointer) error {
	w.WriteInt64(int64(*(*int32)(p)))
	return nil
}

func encodeFloat64(w *Writer, p unsafe.Pointer) error {
	w.WriteFloat64(*(*float64)(p))
	return nil
}

func encodeBool(w *Writer, p unsafe.Pointer) error {
	w.WriteBool(*(*bool)(p))
	return nil
}

// Dynamic encoders

func encodeInterface(w *Writer, p unsafe.Pointer) error {
	val := *(*any)(p)
	if val == nil {
		w.WriteNull()
		return nil
	}

	rv := reflect.ValueOf(val)
	rt := rv.Type()

	enc, err := getEncoder(rt)
	if err != nil {
		return err
	}

	var ptr unsafe.Pointer

	if rt.Kind() == reflect.Pointer {
		ptr = unsafe.Pointer(rv.Pointer())
	} else {
		newPtr := reflect.New(rt)
		newPtr.Elem().Set(rv)
		ptr = unsafe.Pointer(newPtr.Pointer())
	}

	return enc(w, ptr)
}

// Complex encoders
type structFieldEncoder struct {
	offset  uintptr
	encoder EncoderFunc
	key     []byte // preformatted key
}

func compileStructEncoder(t reflect.Type) (EncoderFunc, error) {
	var fields []structFieldEncoder

	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		name := field.Name
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}

		enc, err := compileEncoder(field.Type)
		if err != nil {
			return nil, err
		}

		// Prepare key bytes
		// If it's the first field, we don't add a comma
		// If it's subsequent field, we add a comma
		// We handle the comma logic inside the loop colsure logic below
		// but to keep it branchless in runtime, we usually can't predict
		// which fields are empty (if omitempty).
		// For this MVP, we ignore omitempty and assume standard strict JSON.

		var keyBytes []byte
		if len(fields) > 0 {
			keyBytes = []byte(`,"` + name + `":`)
		} else {
			keyBytes = []byte(`"` + name + `":`)
		}

		fields = append(fields, structFieldEncoder{
			offset:  field.Offset,
			encoder: enc,
			key:     keyBytes,
		})
	}

	return func(w *Writer, p unsafe.Pointer) error {
		w.WriteByte('{')
		for i := range fields {
			f := &fields[i]
			w.Write(f.key)
			fieldPrt := unsafe.Pointer(uintptr(p) + f.offset)
			if err := f.encoder(w, fieldPrt); err != nil {
				return err
			}
		}

		w.WriteByte('}')
		return nil
	}, nil
}

func compileSliceEncoderEnc(t reflect.Type) (EncoderFunc, error) {
	elemType := t.Elem()
	elemSize := elemType.Size()
	elemEnc, err := compileEncoder(elemType)
	if err != nil {
		return nil, err
	}

	return func(w *Writer, p unsafe.Pointer) error {
		header := (*sliceHeader)(p)

		if header.Data == nil && header.Len == 0 {
			w.WriteNull()
			return nil
		}

		w.WriteByte('[')

		for i := range header.Len {
			if i > 0 {
				w.WriteByte(',')
			}

			elemPtr := unsafe.Pointer(uintptr(header.Data) + uintptr(i)*elemSize)
			if err := elemEnc(w, elemPtr); err != nil {
				return err
			}
		}

		w.WriteByte(']')
		return nil
	}, nil
}

func compileMapEncoder(t reflect.Type) (EncoderFunc, error) {
	if t.Key().Kind() != reflect.String {
		return nil, fmt.Errorf("fastjson: maps with %s keys not supported", t.Key().Kind())
	}

	elemEnv, err := compileEncoder(t.Elem())
	if err != nil {
		return nil, err
	}

	return func(w *Writer, p unsafe.Pointer) error {
		mVal := reflect.NewAt(t, p).Elem()
		if mVal.IsNil() {
			w.WriteNull()
			return nil
		}

		w.WriteByte('{')
		iter := mVal.MapRange()
		first := true

		for iter.Next() {
			if !first {
				w.WriteByte(',')
			}
			first = false

			w.WriteStringEscaped(iter.Key().String())
			w.WriteByte(':')

			val := iter.Value()
			var valPtr unsafe.Pointer

			if t.Elem().Kind() == reflect.Pointer {
				valPtr = unsafe.Pointer(val.Pointer())
			} else {
				newPtr := reflect.New(t.Elem())
				newPtr.Elem().Set(val)
				valPtr = unsafe.Pointer(newPtr.Pointer())
			}

			if err := elemEnv(w, valPtr); err != nil {
				return err
			}
		}

		w.WriteByte('}')
		return nil
	}, nil
}
