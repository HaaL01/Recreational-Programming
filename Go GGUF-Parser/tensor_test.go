package gguf

import (
	"bytes"
	"testing"
)

func TestTypeName(t *testing.T) {
	tests := []struct {
		typ  uint32
		want string
	}{
		{0, "F32"},
		{12, "Q4_K"},
		{99, "type99"},
	}
	for _, tt := range tests {
		if got := (Tensor{Type: tt.typ}).TypeName(); got != tt.want {
			t.Errorf("TypeName(%d) = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestElements(t *testing.T) {
	if got := (Tensor{Shape: []uint64{4, 8}}).Elements(); got != 32 {
		t.Errorf("Elements([4 8]) = %d, want 32", got)
	}
}

func TestDataOffsetAlignment(t *testing.T) {
	tests := []struct {
		name  string
		align uint32 // 0 means no general.alignment key
		want  int64
	}{
		{"default", 0, 32},
		{"custom", 64, 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w builder
			if tt.align == 0 {
				w.header(0, 0)
			} else {
				w.header(0, 1)
				w.str("general.alignment")
				w.u32(tUint32)
				w.u32(tt.align)
			}

			f, err := Parse(bytes.NewReader(w.b))
			if err != nil {
				t.Fatal(err)
			}
			if f.DataOffset%tt.want != 0 || f.DataOffset < int64(len(w.b)) {
				t.Errorf("DataOffset = %d for %d-byte header, want next multiple of %d",
					f.DataOffset, len(w.b), tt.want)
			}
		})
	}
}
