package gguf

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	data := validFile()
	f, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if f.Version != 3 || f.KVCount != 3 || f.TensorCount != 1 {
		t.Errorf("header = v%d, %d kvs, %d tensors", f.Version, f.KVCount, f.TensorCount)
	}
	if len(f.Metadata) != 3 {
		t.Errorf("got %d metadata entries, want 3", len(f.Metadata))
	}
	if len(f.Tensors) != 1 || f.Tensors[0].Name != "token_embd.weight" {
		t.Errorf("tensors = %+v", f.Tensors)
	}

	// Data starts at the end of the header, rounded up to 32 bytes.
	want := int64((len(data) + 31) / 32 * 32)
	if f.DataOffset != want {
		t.Errorf("DataOffset = %d, want %d", f.DataOffset, want)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		build func(w *builder)
	}{
		{"bad magic", func(w *builder) {
			w.u32(0x12345678)
		}},
		{"unsupported version", func(w *builder) {
			w.u32(Magic)
			w.u32(1)
		}},
		{"unknown value type", func(w *builder) {
			w.header(0, 1)
			w.str("key")
			w.u32(99)
		}},
		{"string too long", func(w *builder) {
			w.header(0, 1)
			w.u64(maxStringLen + 1)
		}},
		{"too many dims", func(w *builder) {
			w.header(1, 0)
			w.str("x")
			w.u32(maxDims + 1)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w builder
			tt.build(&w)
			if _, err := Parse(bytes.NewReader(w.b)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

// Cutting the file off at any byte must give an error, never a panic.
func TestParseTruncated(t *testing.T) {
	data := validFile()
	for n := 0; n < len(data); n++ {
		if _, err := Parse(bytes.NewReader(data[:n])); err == nil {
			t.Fatalf("no error when truncated to %d of %d bytes", n, len(data))
		}
	}
}

// TestRealModels parses any .gguf files in .models/ and skips if there are none.
func TestRealModels(t *testing.T) {
	paths, _ := filepath.Glob(".models/*.gguf")
	if len(paths) == 0 {
		t.Skip("no models in .models/")
	}
	for _, p := range paths {
		f, err := Open(p)
		if err != nil {
			t.Errorf("%s: %v", p, err)
			continue
		}
		if len(f.Tensors) == 0 {
			t.Errorf("%s: no tensors", p)
		}
	}
}
