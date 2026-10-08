package gguf

import "fmt"

// Tensor describes one tensor. Offset is relative to File.DataOffset.
type Tensor struct {
	Name   string
	Shape  []uint64
	Type   uint32 // ggml type, e.g. 0 = F32, 1 = F16, 2 = Q4_0
	Offset uint64
}

var typeNames = map[uint32]string{
	0: "F32", 1: "F16", 2: "Q4_0", 3: "Q4_1", 6: "Q5_0", 7: "Q5_1",
	8: "Q8_0", 9: "Q8_1", 10: "Q2_K", 11: "Q3_K", 12: "Q4_K", 13: "Q5_K",
	14: "Q6_K", 15: "Q8_K", 16: "IQ2_XXS", 17: "IQ2_XS", 18: "IQ3_XXS",
	19: "IQ1_S", 20: "IQ4_NL", 21: "IQ3_S", 22: "IQ2_S", 23: "IQ4_XS",
	24: "I8", 25: "I16", 26: "I32", 27: "I64", 28: "F64", 29: "IQ1_M",
	30: "BF16", 34: "TQ1_0", 35: "TQ2_0", 39: "MXFP4",
}

// TypeName returns the ggml type name, such as "Q4_K".
func (t Tensor) TypeName() string {
	if name, ok := typeNames[t.Type]; ok {
		return name
	}
	return fmt.Sprintf("type%d", t.Type)
}

// Elements returns the number of values in the tensor.
func (t Tensor) Elements() uint64 {
	n := uint64(1)
	for _, d := range t.Shape {
		n *= d
	}
	return n
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
