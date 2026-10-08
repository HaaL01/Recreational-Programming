// Command ggufdump prints a readable summary of a GGUF model file.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gguf "www.example.com/gguf-parser"
)

// fileTypes names general.file_type values (llama.cpp's llama_ftype).
var fileTypes = map[uint32]string{
	0: "F32", 1: "F16", 2: "Q4_0", 3: "Q4_1", 7: "Q8_0", 8: "Q5_0", 9: "Q5_1",
	10: "Q2_K", 11: "Q3_K_S", 12: "Q3_K_M", 13: "Q3_K_L", 14: "Q4_K_S",
	15: "Q4_K_M", 16: "Q5_K_S", 17: "Q5_K_M", 18: "Q6_K", 19: "IQ2_XXS",
	20: "IQ2_XS", 21: "Q2_K_S", 22: "IQ3_XS", 23: "IQ3_XXS", 24: "IQ1_S",
	25: "IQ4_NL", 26: "IQ3_S", 27: "IQ3_M", 28: "IQ2_S", 29: "IQ2_M",
	30: "IQ4_XS", 31: "IQ1_M", 32: "BF16", 36: "TQ1_0", 37: "TQ2_0",
}

func main() {
	listTensors := flag.Bool("t", false, "list every tensor")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: ggufdump [-t] <file.gguf>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	path := flag.Arg(0)

	f, err := gguf.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	printSummary(f, path)
	printMetadata(f)
	printTensors(f, *listTensors)
}

func printSummary(f *gguf.File, path string) {
	arch, _ := f.Get("general.architecture")
	name, _ := f.Get("general.name")
	fmt.Printf("%v (%v)\n", name, arch)

	size := ""
	if st, err := os.Stat(path); err == nil {
		size = fmt.Sprintf(", %.2f GiB", float64(st.Size())/(1<<30))
	}
	row("File", fmt.Sprintf("%s%s, GGUF v%d", filepath.Base(path), size, f.Version))

	if v, ok := f.Get("general.file_type"); ok {
		if ft, ok := v.(uint32); ok && fileTypes[ft] != "" {
			row("Quantization", fileTypes[ft])
		}
	}

	var params uint64
	for _, t := range f.Tensors {
		params += t.Elements()
	}
	row("Parameters", humanCount(params))

	for _, k := range []struct{ label, key string }{
		{"Context", "context_length"},
		{"Layers", "block_count"},
		{"Embedding", "embedding_length"},
		{"Heads", "attention.head_count"},
		{"KV heads", "attention.head_count_kv"},
	} {
		if v, ok := f.Get(fmt.Sprintf("%v.%s", arch, k.key)); ok {
			row(k.label, fmt.Sprint(v))
		}
	}
	fmt.Println()
}

func printMetadata(f *gguf.File) {
	fmt.Printf("Metadata (%d keys)\n", len(f.Metadata))
	width := 0
	for _, kv := range f.Metadata {
		width = max(width, len(kv.Key))
	}
	for _, kv := range f.Metadata {
		fmt.Printf("  %-*s  %s\n", width, kv.Key, short(kv.Value))
	}
	fmt.Println()
}

func printTensors(f *gguf.File, all bool) {
	fmt.Printf("Tensors (%d, data at offset %d)\n", len(f.Tensors), f.DataOffset)
	if all {
		for _, t := range f.Tensors {
			fmt.Printf("  %-40s %-7s %v\n", t.Name, t.TypeName(), t.Shape)
		}
		return
	}

	// Count tensors and parameters per type, in order of first appearance.
	var order []string
	count := map[string]int{}
	elems := map[string]uint64{}
	for _, t := range f.Tensors {
		name := t.TypeName()
		if count[name] == 0 {
			order = append(order, name)
		}
		count[name]++
		elems[name] += t.Elements()
	}
	for _, name := range order {
		fmt.Printf("  %-7s %4d tensors  %8s params\n", name, count[name], humanCount(elems[name]))
	}
	fmt.Println("  (use -t to list every tensor)")
}

func row(label, value string) {
	fmt.Printf("  %-13s %s\n", label, value)
}

// short formats a metadata value on one line, abbreviating long ones.
func short(v any) string {
	switch v := v.(type) {
	case []any:
		if len(v) > 8 {
			return fmt.Sprintf("[%d items]", len(v))
		}
	case string:
		first, _, multiline := strings.Cut(v, "\n")
		if multiline || len(v) > 60 {
			return fmt.Sprintf("%.60s... (%d chars)", first, len(v))
		}
	}
	return fmt.Sprint(v)
}

func humanCount(n uint64) string {
	switch {
	case n >= 1e9:
		return fmt.Sprintf("%.2fB", float64(n)/1e9)
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1e3:
		return fmt.Sprintf("%.1fK", float64(n)/1e3)
	}
	return fmt.Sprint(n)
}
