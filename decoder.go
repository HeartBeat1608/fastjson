package fastjson

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

// DecoderFunc is the signature of our compiled decoders.
// p is the unsafe pointer to the value we want to decode into.
type DecoderFunc func(it *Iterator, p unsafe.Pointer) error

// cache stores compiled decoders to avoid repeated reflection analysis.
var decoderCache sync.Map

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// getDecoder returns a cached decoder or compiles a new one.
func getDecoder(t reflect.Type) (DecoderFunc, error) {
	if f, ok := decoderCache.Load(t); ok {
		return f.(DecoderFunc), nil
	}

	// Compile new decoder for this type.
	dec, err := compileDecoder(t)
	if err != nil {
		return nil, err
	}

	decoderCache.Store(t, dec)
	return dec, nil
}

// compileDecoder switches on the type to return the correct primitive or struct decoder.
func compileDecoder(t reflect.Type) (DecoderFunc, error) {
	switch t.Kind() {
	case reflect.String:
		return decodeString, nil
	case reflect.Int, reflect.Int64:
		return decodeInt64, nil
	case reflect.Int32:
		return decodeInt32, nil
	case reflect.Float64:
		return decodeFloat64, nil
	case reflect.Bool:
		return decodeBool, nil
	case reflect.Interface:
		return decodeInterface, nil
	case reflect.Struct:
		return compileStructDecoder(t)
	case reflect.Slice:
		return compileSliceDecoder(t)
	case reflect.Map:
		return compileMapDecoder(t)
	case reflect.Pointer:
		elemDec, err := compileDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return func(it *Iterator, p unsafe.Pointer) error {
			ptrVal := *(*unsafe.Pointer)(p)
			if ptrVal == nil {
				newVal := reflect.New(t.Elem())
				ptrVal = unsafe.Pointer(newVal.Pointer())
				*(*unsafe.Pointer)(p) = ptrVal
			}
			return elemDec(it, ptrVal)
		}, nil
	default:
		return nil, fmt.Errorf("fastjson: unsupported type: %s", t.Kind())
	}
}

// Primitive decoders
func decodeString(it *Iterator, p unsafe.Pointer) error {
	s, err := it.ReadString()
	if err != nil {
		return err
	}
	*(*string)(p) = s
	return nil
}

func decodeInt64(it *Iterator, p unsafe.Pointer) error {
	i, err := it.ReadInt64()
	if err != nil {
		return err
	}
	*(*int)(p) = int(i)
	return nil
}

func decodeInt32(it *Iterator, p unsafe.Pointer) error {
	i, err := it.ReadInt64()
	if err != nil {
		return err
	}
	*(*int32)(p) = int32(i)
	return nil
}

func decodeFloat64(it *Iterator, p unsafe.Pointer) error {
	f, err := it.ReadFloat64()
	if err != nil {
		return err
	}
	*(*float64)(p) = f
	return nil
}

func decodeBool(it *Iterator, p unsafe.Pointer) error {
	b, err := it.ReadBool()
	if err != nil {
		return err
	}
	*(*bool)(p) = b
	return nil
}

// Complex decoders

// compileStructDecoder handles []T
func compileSliceDecoder(t reflect.Type) (DecoderFunc, error) {
	elemType := t.Elem()
	elemSize := elemType.Size()
	elemDec, err := compileDecoder(elemType)
	if err != nil {
		return nil, err
	}

	// sliceType := t

	return func(it *Iterator, p unsafe.Pointer) error {
		header := (*sliceHeader)(p)

		if err := it.ReadArrayStart(); err != nil {
			return err
		}

		header.Len = 0

		for {
			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == ']' {
				it.head++
				return nil
			}

			if header.Len >= header.Cap {
				newCap := header.Cap * 2
				if newCap == 0 {
					newCap = 8
				}

				src := reflect.NewAt(t, p).Elem()
				newSlice := reflect.MakeSlice(t, header.Len, newCap)
				reflect.Copy(newSlice, src)
				src.Set(newSlice)

				header = (*sliceHeader)(p)
			}

			elemPtr := unsafe.Pointer(uintptr(header.Data) + uintptr(header.Len)*elemSize)

			if err := elemDec(it, elemPtr); err != nil {
				return err
			}
			header.Len++

			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == ',' {
				it.head++
				continue
			} else if it.head < it.dataLen && it.data[it.head] == ']' {
				it.head++
				return nil
			} else {
				return it.error("expected ',' or ']'")
			}
		}
	}, nil
}

// compileStructDecoder handles map[string]T
func compileMapDecoder(t reflect.Type) (DecoderFunc, error) {
	keyType := t.Key()
	if keyType.Kind() != reflect.String {
		return nil, fmt.Errorf("fastjson: maps with %s keys not supported", keyType.Kind())
	}

	elemType := t.Elem()
	elemDec, err := compileDecoder(elemType)
	if err != nil {
		return nil, err
	}

	mapType := t

	return func(it *Iterator, p unsafe.Pointer) error {
		if err := it.ReadObjectStart(); err != nil {
			return err
		}

		mapVal := reflect.NewAt(mapType, p).Elem()

		if mapVal.IsNil() {
			mapVal.Set(reflect.MakeMap(mapType))
		}

		for {
			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == '}' {
				it.head++
				return nil
			}

			key, err := it.ReadString()
			if err != nil {
				return err
			}

			if err := it.ReadColon(); err != nil {
				return err
			}

			// Create a new value for the element
			// We allocate a new one because maps store pointers/copies internally
			newElem := reflect.New(elemType) // returns *T

			if err := elemDec(it, unsafe.Pointer(newElem.Pointer())); err != nil {
				return err
			}

			mapVal.SetMapIndex(reflect.ValueOf(key), newElem.Elem())
			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == ',' {
				it.head++
				continue
			} else if it.head < it.dataLen && it.data[it.head] == '}' {
				it.head++
				return nil
			} else {
				return it.error("expected ',' or '}'")
			}
		}
	}, nil
}

type fieldInfo struct {
	offset  uintptr
	decoder DecoderFunc
}

func compileStructDecoder(t reflect.Type) (DecoderFunc, error) {
	fieldMap := make(map[string]*fieldInfo)

	for i := 0; i < t.NumField(); i++ {
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

		dec, err := compileDecoder(field.Type)
		if err != nil {
			return nil, err
		}

		fieldMap[name] = &fieldInfo{
			offset:  field.Offset,
			decoder: dec,
		}
	}

	return func(it *Iterator, p unsafe.Pointer) error {
		if err := it.ReadObjectStart(); err != nil {
			return err
		}

		for {
			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == '}' {
				it.head++
				return nil
			}

			key, err := it.ReadString()
			if err != nil {
				return err
			}

			if err := it.ReadColon(); err != nil {
				return err
			}

			it.skipWhiteSpace()

			if info, ok := fieldMap[key]; ok {
				fieldPtr := unsafe.Pointer(uintptr(p) + info.offset)
				if err := info.decoder(it, fieldPtr); err != nil {
					return err
				}
			} else {
				if err := it.SkipValue(); err != nil {
					return err
				}
			}

			it.skipWhiteSpace()
			if it.head < it.dataLen && it.data[it.head] == ',' {
				it.head++
				continue
			} else if it.head < it.dataLen && it.data[it.head] == '}' {
				it.head++
				return nil
			} else {
				return it.error("expected ',' or '}'")
			}
		}
	}, nil
}

func decodeInterface(it *Iterator, p unsafe.Pointer) error {
	val, err := readValue(it)
	if err != nil {
		return err
	}
	*(*any)(p) = val
	return nil
}

func readValue(it *Iterator) (any, error) {
	it.skipWhiteSpace()
	if it.head >= it.dataLen {
		return nil, it.error("unexpected end of JSON")
	}

	c := it.data[it.head]
	switch c {
	case '"':
		return it.ReadString()
	case '{':
		return readGenericMap(it)
	case '[':
		return readGenericSlice(it)
	case 't', 'f':
		return it.ReadBool()
	case 'n':
		return nil, it.ReadNull()
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return it.ReadFloat64()
	default:
		return nil, it.error(fmt.Sprintf("unexpected character: %c", c))
	}
}

func readGenericMap(it *Iterator) (any, error) {
	m := make(map[string]any)
	if err := it.ReadObjectStart(); err != nil {
		return nil, err
	}

	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == '}' {
		it.head++
		return m, nil
	}

	for {
		key, err := it.ReadString()
		if err != nil {
			return nil, err
		}

		if err := it.ReadColon(); err != nil {
			return nil, err
		}

		val, err := readValue(it)
		if err != nil {
			return nil, err
		}

		m[key] = val

		it.skipWhiteSpace()
		if it.head < it.dataLen && it.data[it.head] == ',' {
			it.head++
			continue
		} else if it.head < it.dataLen && it.data[it.head] == '}' {
			it.head++
			return m, nil
		} else {
			return nil, it.error("expected ',' or '}'")
		}
	}
}

func readGenericSlice(it *Iterator) (any, error) {
	l := make([]any, 0, 8) // slight pre-allocation
	if err := it.ReadArrayStart(); err != nil {
		return nil, err
	}

	it.skipWhiteSpace()
	if it.head < it.dataLen && it.data[it.head] == ']' {
		it.head++
		return l, nil
	}

	for {
		val, err := readValue(it)
		if err != nil {
			return nil, err
		}

		l = append(l, val)

		it.skipWhiteSpace()
		if it.head < it.dataLen && it.data[it.head] == ',' {
			it.head++
			continue
		} else if it.head < it.dataLen && it.data[it.head] == ']' {
			it.head++
			return l, nil
		} else {
			return nil, it.error("expected ',' or ']'")
		}
	}
}
