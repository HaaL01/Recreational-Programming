package gguf

import (
	"encoding/binary"
	"fmt"
	"io"
)

// reader decodes little-endian values from a stream. It tracks how many bytes
// have been consumed and keeps the first error; once an error is set, every
// later read is a no-op returning zero, so callers check err once at the end.
type reader struct {
	r   io.Reader
	pos int64
	err error
	buf [8]byte
}

// fill reads exactly len(p) bytes into p.
func (r *reader) fill(p []byte) {
	if r.err != nil {
		return
	}
	n, err := io.ReadFull(r.r, p)
	r.pos += int64(n)
	if err != nil {
		r.err = fmt.Errorf("at offset %d: %w", r.pos, err)
	}
}

// scalar reads n bytes (n <= 8) into the scratch buffer.
func (r *reader) scalar(n int) []byte {
	b := r.buf[:n]
	clear(b)
	r.fill(b)
	return b
}

func (r *reader) u32() uint32 { return binary.LittleEndian.Uint32(r.scalar(4)) }
func (r *reader) u64() uint64 { return binary.LittleEndian.Uint64(r.scalar(8)) }
