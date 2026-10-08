package gguf

import "fmt"

// KV is one metadata entry. Value is a uint8, int8, uint16, int16, uint32,
// int32, uint64, int64, float32, float64, bool, string, or []any for arrays.
type KV struct {
	Key   string
	Value any
}

// Get returns the value stored under key.
func (f *File) Get(key string) (any, bool) {
	for _, kv := range f.Metadata {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return nil, false
}

func (r *reader) kv() KV {
	key := r.str()
	return KV{Key: key, Value: r.value(r.u32())}
}

// value reads one value of GGUF type t.
func (r *reader) value(t uint32) any {
	switch t {
	case 0:
		return r.u8()
	case 1:
		return r.i8()
	case 2:
		return r.u16()
	case 3:
		return r.i16()
	case 4:
		return r.u32()
	case 5:
		return r.i32()
	case 6:
		return r.f32()
	case 7:
		return r.bool()
	case 8:
		return r.str()
	case 9:
		return r.array()
	case 10:
		return r.u64()
	case 11:
		return r.i64()
	case 12:
		return r.f64()
	}
	if r.err == nil {
		r.err = fmt.Errorf("at offset %d: unknown value type %d", r.pos, t)
	}
	return nil
}

// array reads an element type, a uint64 count, then that many values.
func (r *reader) array() []any {
	t := r.u32()
	n := r.u64()
	var s []any
	for i := uint64(0); i < n && r.err == nil; i++ {
		s = append(s, r.value(t))
	}
	return s
}
