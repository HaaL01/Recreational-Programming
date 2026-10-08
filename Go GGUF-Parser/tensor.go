package gguf

import "fmt"

// Tensor describes one tensor. Offset is relative to File.DataOffset.
type Tensor struct {
	Name   string
	Shape  []uint64
	Type   uint32 // ggml type, e.g. 0 = F32, 1 = F16, 2 = Q4_0
	Offset uint64
}

// maxDims matches GGML_MAX_DIMS.
const maxDims = 4

func (r *reader) tensor() Tensor {
	t := Tensor{Name: r.str()}
	n := r.u32()
	if r.err == nil && n > maxDims {
		r.err = fmt.Errorf("at offset %d: tensor %q has %d dims", r.pos, t.Name, n)
		return t
	}
	for i := uint32(0); i < n; i++ {
		t.Shape = append(t.Shape, r.u64())
	}
	t.Type = r.u32()
	t.Offset = r.u64()
	return t
}
