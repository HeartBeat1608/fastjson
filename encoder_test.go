package fastjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

// --- Test Data Models ---

type Primitives struct {
	S   string  `json:"s"`
	I   int     `json:"i"`
	I64 int64   `json:"i64"`
	F   float64 `json:"f"`
	B   bool    `json:"b"`
}

type Nested struct {
	Title string     `json:"title"`
	Prim  Primitives `json:"prim"`
}

type Pointers struct {
	PStr *string `json:"p_str"`
	PInt *int    `json:"p_int"`
	PNil *int    `json:"p_nil"`
}

type Collections struct {
	List []string          `json:"list"`
	Dict map[string]string `json:"dict"`
}

type Generic struct {
	Data any `json:"data"`
}

// --- Tests ---

func TestMarshal_Primitives(t *testing.T) {
	input := Primitives{S: "Hello \"World\"", I: -42, I64: 1234567890123, F: 3.14159, B: true}

	// 1. Marshal with FastJSON
	data, err := Marshal(&input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// 2. Unmarshal with Standard Lib to verify validity
	var output Primitives
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Generated invalid JSON: %s, error: %v", string(data), err)
	}

	// 3. Compare
	if !reflect.DeepEqual(input, output) {
		t.Errorf("Mismatch.\nInput:  %+v\nOutput: %+v", input, output)
	}
}

func TestMarshal_Collections(t *testing.T) {
	input := Collections{
		List: []string{"one", "two", "three"},
		Dict: map[string]string{"key1": "value1", "key2": "value2"},
	}

	data, err := Marshal(&input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Note: Map keys might be out of order, so string comparison is flaky.
	// We decode back to verify content.
	var output Collections
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Generated invalid JSON: %s, error: %v", string(data), err)
	}

	if !reflect.DeepEqual(input, output) {
		t.Errorf("Mismatch.\nInput:  %+v\nOutput: %+v", input, output)
	}
}

func TestMarshal_Interfaces(t *testing.T) {
	// Interface holding a map
	input := Generic{
		Data: map[string]int{"a": 1, "b": 2},
	}

	data, err := Marshal(&input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify
	var output Generic
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Generated invalid JSON: %s", string(data))
	}

	// DeepEqual with interface{} containing maps is tricky because JSON numbers become float64.
	// We check manually.
	m, ok := output.Data.(map[string]any)
	if !ok {
		t.Fatalf("Expected map in interface, got %T", output.Data)
	}
	if m["a"] != 1.0 || m["b"] != 2.0 {
		t.Errorf("Interface map content mismatch: %v", m)
	}
}

func TestMarshal_MixedTypes(t *testing.T) {
	// A complex scenario mixing pointers, nil, and nested structs
	str := "ptr_val"
	input := Pointers{
		PStr: &str,
		PInt: nil,
		PNil: nil,
	}

	data, err := Marshal(&input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var output Pointers
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("Generated invalid JSON: %s", string(data))
	}

	if *output.PStr != "ptr_val" {
		t.Errorf("Pointer string mismatch")
	}
	if output.PInt != nil {
		t.Errorf("Expected nil int pointer")
	}
}

func TestMarshal_AnonymousStructs(t *testing.T) {
	input := struct {
		AnonID int `json:"anon_id"`
	}{
		AnonID: 999,
	}

	data, err := Marshal(&input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"anon_id":999}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}
