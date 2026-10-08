package gguf

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

func (r *reader) u8() uint8   { return r.scalar(1)[0] }
func (r *reader) u16() uint16 { return binary.LittleEndian.Uint16(r.scalar(2)) }
func (r *reader) u32() uint32 { return binary.LittleEndian.Uint32(r.scalar(4)) }
func (r *reader) u64() uint64 { return binary.LittleEndian.Uint64(r.scalar(8)) }

func (r *reader) i8() int8   { return int8(r.u8()) }
func (r *reader) i16() int16 { return int16(r.u16()) }
func (r *reader) i32() int32 { return int32(r.u32()) }
func (r *reader) i64() int64 { return int64(r.u64()) }

func (r *reader) f32() float32 { return math.Float32frombits(r.u32()) }
func (r *reader) f64() float64 { return math.Float64frombits(r.u64()) }
func (r *reader) bool() bool   { return r.u8() != 0 }

// maxStringLen bounds a single string so a corrupt length can't force a huge
// allocation. Real files top out around a few hundred KiB (chat templates).
const maxStringLen = 1 << 24

// str reads a GGUF string: a uint64 byte length followed by UTF-8 bytes.
func (r *reader) str() string {
	n := r.u64()
	if r.err != nil {
		return ""
	}
	if n > maxStringLen {
		r.err = fmt.Errorf("at offset %d: string length %d exceeds limit", r.pos, n)
		return ""
	}
	b := make([]byte, n)
	r.fill(b)
	return string(b)
}
