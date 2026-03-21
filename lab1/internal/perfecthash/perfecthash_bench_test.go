package perfecthash

import "testing"

func BenchmarkPerfectHashBuild(b *testing.B) {
	const n = 100000
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := &Builder{}
		if _, err := builder.Build(keys); err != nil {
			b.Fatalf("Build error: %v", err)
		}
	}
}

func BenchmarkPerfectHashLookup(b *testing.B) {
	const n = 100000
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		keys[i] = []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
	}
	builder := &Builder{}
	table, err := builder.Build(keys)
	if err != nil {
		b.Fatalf("Build error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := keys[i%len(keys)]
		if _, ok := table.Lookup(k); !ok {
			b.Fatalf("Lookup failed for existing key")
		}
	}
}

