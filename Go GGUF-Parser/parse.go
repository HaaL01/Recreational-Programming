package gguf

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Parse reads a GGUF header from r.
func Parse(r io.Reader) (*File, error) {
	rd := &reader{r: r}

	magic := rd.u32()
	if rd.err != nil {
		return nil, rd.err
	}
	if magic != Magic {
		return nil, fmt.Errorf("not a GGUF file (magic %#08x)", magic)
	}

	f := &File{}
	f.Version = rd.u32()
	if rd.err == nil && f.Version != 2 && f.Version != 3 {
		return nil, fmt.Errorf("unsupported GGUF version %d", f.Version)
	}
	f.TensorCount = rd.u64()
	f.KVCount = rd.u64()

	if rd.err != nil {
		return nil, rd.err
	}
	return f, nil
}

// Open parses the GGUF file at path.
func Open(path string) (*File, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return Parse(bufio.NewReader(fh))
}
