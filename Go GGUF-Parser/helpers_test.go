package gguf

import "encoding/binary"

// builder assembles GGUF bytes by hand so tests don't need a model file.
type builder struct{ b []byte }

func (w *builder) u8(v uint8)   { w.b = append(w.b, v) }
func (w *builder) u16(v uint16) { w.b = binary.LittleEndian.AppendUint16(w.b, v) }
func (w *builder) u32(v uint32) { w.b = binary.LittleEndian.AppendUint32(w.b, v) }
func (w *builder) u64(v uint64) { w.b = binary.LittleEndian.AppendUint64(w.b, v) }
func (w *builder) str(s string) { w.u64(uint64(len(s))); w.b = append(w.b, s...) }

// header writes the magic, version 3, and the tensor and KV counts.
func (w *builder) header(tensors, kvs uint64) {
	w.u32(Magic)
	w.u32(3)
	w.u64(tensors)
	w.u64(kvs)
}

// GGUF value type codes, named here only to make test data readable.
const (
	tUint8 = iota
	tInt8
	tUint16
	tInt16
	tUint32
	tInt32
	tFloat32
	tBool
	tString
	tArray
	tUint64
	tInt64
	tFloat64
)

// validFile returns a small GGUF file: 3 metadata keys and 1 tensor.
func validFile() []byte {
	var w builder
	w.header(1, 3)

	w.str("general.architecture")
	w.u32(tString)
	w.str("llama")

	w.str("general.alignment")
	w.u32(tUint32)
	w.u32(32)

	w.str("tokenizer.ggml.tokens")
	w.u32(tArray)
	w.u32(tString)
	w.u64(2)
	w.str("a")
	w.str("b")

	w.str("token_embd.weight")
	w.u32(2) // dims
	w.u64(4)
	w.u64(8)
	w.u32(0) // F32
	w.u64(0) // offset

	return w.b
}
