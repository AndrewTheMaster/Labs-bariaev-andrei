package main

import (
	"fmt"
	"os"
	"sort"

	"siaod-hw1/internal/lshtext"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: lsh-demo file1.txt [file2.txt ...]")
		os.Exit(1)
	}

	idx := lshtext.NewIndex(3, 64, 8, 8)

	// Индексируем все переданные файлы.
	for i, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		if err := idx.Add(i, data); err != nil {
			panic(err)
		}
	}

	// Ищем пары дублей.
	pairs, err := idx.FullScanDuplicates(0.5)
	if err != nil {
		panic(err)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Similarity > pairs[j].Similarity
	})

	for _, p := range pairs {
		fmt.Printf("pair (%d, %d) similarity=%.3f\n", p.ID1, p.ID2, p.Similarity)
	}
}

