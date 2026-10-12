package gguf

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

func TestValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		typ   uint32
		value func(w *builder)
		want  any
	}{
		{"uint8", tUint8, func(w *builder) { w.u8(7) }, uint8(7)},
		{"int8", tInt8, func(w *builder) { w.u8(0xFF) }, int8(-1)},
		{"uint16", tUint16, func(w *builder) { w.u16(500) }, uint16(500)},
		{"int16", tInt16, func(w *builder) { w.u16(0xFFFF) }, int16(-1)},
		{"uint32", tUint32, func(w *builder) { w.u32(70000) }, uint32(70000)},
		{"int32", tInt32, func(w *builder) { w.u32(0xFFFFFFFF) }, int32(-1)},
		{"float32", tFloat32, func(w *builder) { w.u32(math.Float32bits(1.5)) }, float32(1.5)},
		{"bool", tBool, func(w *builder) { w.u8(1) }, true},
		{"string", tString, func(w *builder) { w.str("hi") }, "hi"},
		{"uint64", tUint64, func(w *builder) { w.u64(1 << 40) }, uint64(1 << 40)},
		{"int64", tInt64, func(w *builder) { w.u64(math.MaxUint64) }, int64(-1)},
		{"float64", tFloat64, func(w *builder) { w.u64(math.Float64bits(2.5)) }, 2.5},
		{"array", tArray, func(w *builder) {
			w.u32(tInt32)
			w.u64(2)
			w.u32(1)
			w.u32(2)
		}, []any{int32(1), int32(2)}},
		{"nested array", tArray, func(w *builder) {
			w.u32(tArray)
			w.u64(1)
			w.u32(tBool)
			w.u64(1)
			w.u8(1)
		}, []any{[]any{true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w builder
			w.header(0, 1)
			w.str("key")
			w.u32(tt.typ)
			tt.value(&w)

			f, err := Parse(bytes.NewReader(w.b))
			if err != nil {
				t.Fatal(err)
			}
			if got, _ := f.Get("key"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	f, err := Parse(bytes.NewReader(validFile()))
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := f.Get("general.architecture"); !ok || v != "llama" {
		t.Errorf("Get(general.architecture) = %v, %v", v, ok)
	}
	if _, ok := f.Get("missing"); ok {
		t.Error("Get(missing) reported ok")
	}
}
