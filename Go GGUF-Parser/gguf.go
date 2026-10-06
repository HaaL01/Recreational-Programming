// Package gguf parses the header of GGUF model files: metadata and tensor
// layout. Tensor data itself is never read.
package gguf

// Magic is the bytes "GGUF" read as a little-endian uint32.
const Magic = 0x46554747

// File is the parsed header of a GGUF file.
type File struct {
	Version     uint32
	TensorCount uint64
	KVCount     uint64
}
