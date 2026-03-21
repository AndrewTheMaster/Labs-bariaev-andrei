package hashfs

import (
	"os"
	"path/filepath"
	"testing"

	"siaod-hw1/internal/gen"
)

func TestStore_BasicCRUD(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.dat")

	store, err := Open(path, Options{
		BucketCount: 1 << 16,
	})
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	const n = 5000

	type kv struct {
		key   []byte
		value []byte
	}

	data := make([]kv, 0, n)
	ref := make(map[string][]byte, n)

	for i := 0; i < n; i++ {
		k := gen.RandomBytes(16)
		v := gen.RandomBytes(32)
		if err := store.Put(k, v); err != nil {
			t.Fatalf("Put error: %v", err)
		}
		data = append(data, kv{key: k, value: v})
		ref[string(k)] = v
	}

	for _, item := range data {
		got, err := store.Get(item.key)
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}
		if !equalBytes(got, item.value) {
			t.Fatalf("Get mismatch: expected %x, got %x", item.value, got)
		}
	}

	// Обновим половину ключей.
	for i := 0; i < n; i += 2 {
		newVal := gen.RandomBytes(48)
		if err := store.Put(data[i].key, newVal); err != nil {
			t.Fatalf("Put(update) error: %v", err)
		}
		data[i].value = newVal
		ref[string(data[i].key)] = newVal
	}

	// Удалим каждую третью запись.
	for i := 0; i < n; i += 3 {
		if err := store.Delete(data[i].key); err != nil {
			t.Fatalf("Delete error: %v", err)
		}
		delete(ref, string(data[i].key))
	}

	// Проверим соответствие с эталонной картой.
	for _, item := range data {
		val, ok := ref[string(item.key)]
		got, err := store.Get(item.key)
		if !ok {
			if err == nil {
				t.Fatalf("expected not found for deleted key, got value %x", got)
			}
			if err != ErrNotFound {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Get after update/delete error: %v", err)
		}
		if !equalBytes(got, val) {
			t.Fatalf("value mismatch after update/delete: expected %x, got %x", val, got)
		}
	}
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.dat")

	opts := Options{
		BucketCount: 1 << 14,
	}

	s, err := Open(path, opts)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}

	k := []byte("key-1")
	v := []byte("value-1")
	if err := s.Put(k, v); err != nil {
		t.Fatalf("Put error: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Проверим, что файл действительно существует и не пустой.
	if info, err := os.Stat(path); err != nil {
		t.Fatalf("Stat error: %v", err)
	} else if info.Size() == 0 {
		t.Fatalf("expected non-empty file size")
	}

	s2, err := Open(path, opts)
	if err != nil {
		t.Fatalf("Open(second) error: %v", err)
	}
	defer s2.Close()

	got, err := s2.Get(k)
	if err != nil {
		t.Fatalf("Get after reopen error: %v", err)
	}
	if !equalBytes(got, v) {
		t.Fatalf("persisted value mismatch: expected %s, got %s", v, got)
	}
}

func TestStore_EdgeCases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "edge.dat")

	store, err := Open(path, Options{
		BucketCount: 1 << 10,
		MaxValueSize: 64,
	})
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	// Пустой ключ допустим.
	if err := store.Put([]byte{}, []byte("empty-key")); err != nil {
		t.Fatalf("Put empty key error: %v", err)
	}
	if got, err := store.Get([]byte{}); err != nil || !equalBytes(got, []byte("empty-key")) {
		t.Fatalf("Get empty key mismatch: got=%q err=%v", got, err)
	}

	// Value больше лимита должно приводить к ошибке.
	tooLarge := make([]byte, 65)
	if err := store.Put([]byte("k"), tooLarge); err == nil {
		t.Fatalf("expected value-too-large error")
	}

	// Удаление несуществующего ключа не должно ломать структуру.
	if err := store.Delete([]byte("missing-key")); err != nil {
		t.Fatalf("Delete missing key error: %v", err)
	}
	if _, err := store.Get([]byte("missing-key")); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after deleting missing key, got %v", err)
	}
}

func TestStore_MultiReopenAndRandomOps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi-reopen.dat")
	opts := Options{BucketCount: 1 << 12}

	r := gen.NewDeterministicSource(42)
	ref := make(map[string][]byte)

	for round := 0; round < 3; round++ {
		store, err := Open(path, opts)
		if err != nil {
			t.Fatalf("Open round=%d error: %v", round, err)
		}

		for i := 0; i < 1000; i++ {
			key := make([]byte, 8)
			val := make([]byte, 16)
			for j := range key {
				key[j] = byte(r.Intn(256))
			}
			for j := range val {
				val[j] = byte(r.Intn(256))
			}

			switch r.Intn(3) {
			case 0:
				if err := store.Put(key, val); err != nil {
					t.Fatalf("Put error: %v", err)
				}
				ref[string(key)] = append([]byte(nil), val...)
			case 1:
				if err := store.Delete(key); err != nil {
					t.Fatalf("Delete error: %v", err)
				}
				delete(ref, string(key))
			default:
				got, err := store.Get(key)
				exp, ok := ref[string(key)]
				if !ok {
					if err != ErrNotFound {
						t.Fatalf("expected ErrNotFound, got %v", err)
					}
					continue
				}
				if err != nil || !equalBytes(got, exp) {
					t.Fatalf("Get mismatch err=%v", err)
				}
			}
		}

		if err := store.Close(); err != nil {
			t.Fatalf("Close error: %v", err)
		}
	}
}

