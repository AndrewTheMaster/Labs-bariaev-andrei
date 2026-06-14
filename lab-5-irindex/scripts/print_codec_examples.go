//go:build ignore

package main

import (
	"log"
	"os"

	"siaod-hw5-irindex/internal/ir"
)

func main() {
	out := "metrics/raw/codec_examples.tsv"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll("metrics/raw", 0o755); err != nil {
		log.Fatal(err)
	}
	if err := ir.WriteCodecExamplesTSV(out); err != nil {
		log.Fatal(err)
	}
}
