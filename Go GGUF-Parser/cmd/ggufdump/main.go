// Command ggufdump prints the header of a GGUF model file.
package main

import (
	"fmt"
	"os"

	gguf "www.example.com/gguf-parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: ggufdump <file.gguf>")
		os.Exit(2)
	}

	f, err := gguf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("Version:  %d\n", f.Version)
	fmt.Printf("Tensors:  %d\n", f.TensorCount)
	fmt.Printf("Metadata: %d keys\n", f.KVCount)
}
