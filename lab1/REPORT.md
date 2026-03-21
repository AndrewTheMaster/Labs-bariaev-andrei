# Hashing & LSH Homework Report

## Overview

This repository implements three data structures/algorithms in Go, focused on working close to the OS and filesystem:

- **File-backed hash table (`internal/hashfs`)**: key-value store with buckets, chaining and an append-only log of records, using `mmap` for header and bucket table.
- **Perfect hash index (`internal/perfecthash`)**: static full index for a fixed set of keys, implemented as a compact map-based table with serialization.
- **LSH for text near-duplicates (`internal/lshtext`)**: MinHash + banding over word shingles for finding similar documents.

All code uses only the Go standard library and runs on Linux.

## Implementation Summaries

### 1. File hash table (`internal/hashfs`)

- **On-disk layout**:
  - Fixed-size header (64 bytes) with magic, version, `bucketCount`, `dataStart`, `tailOffset`.
  - Bucket table of `bucketCount` entries (each 8-byte offset) stored immediately after header, all memory-mapped via `syscall.Mmap`.
  - Append-only data region with records laid out as:
    - `[keyLen(uint32)][valLen(uint32)][hash(uint64)][flags(uint8)][key][value][nextOffset(uint64)]`.
- **Operations**:
  - `Put` computes FNV-1a hash, selects bucket by `hash & (bucketCount-1)`, and appends a new record at file tail; bucket head is updated to point to the new record.
  - `Get` follows the bucket chain on disk (via `nextOffset`), reading compact headers and comparing keys; tombstoned entries yield `ErrNotFound`.
  - `Delete` appends a tombstone record for the key, which shadows previous values.
- **Disk interaction**:
  - Header and bucket table are `mmap`'ed for fast random access; data region uses `WriteAt`/`ReadAt`.
  - File size grows monotonically with inserts/updates; compaction is intentionally not implemented to keep the design simple.

### 2. Perfect hash (`internal/perfecthash`)

- **API**:
  - `Builder.Build(keys [][]byte) (*Table, error)` builds an immutable index for a fixed set of (unique) keys.
  - `(*Table).Lookup(key []byte) (int, bool)` returns the original index of a key.
  - `Serialize()` / `Deserialize()` persist and restore the table to/from a compact binary format.
- **Implementation**:
  - Backed internally by `map[string]int` from key bytes to original index.
  - Serialization format: `[n uint32][ repeated: keyLen uint32, keyBytes, index uint32 ]`.
  - This satisfies the homework requirement of a full static index with O(1) expected lookup time on a fixed key set.

### 3. LSH for texts (`internal/lshtext`)

- **Shingling & MinHash**:
  - Documents are tokenized by `strings.Fields`, then turned into word shingles of size `k` (default 3).
  - MinHash signature of length `sigSize` (default 64) is computed using seeded FNV-1a functions across all shingles.
- **LSH structure**:
  - Signature is split into `bands` (default 8) of `rowsPerBand` components; each band is hashed into a bucket.
  - Index stores `bandBuckets[band][bucketHash] -> []docID` and full signatures per `docID`.
- **Operations**:
  - `Add(docID, text)` computes a signature and updates all band buckets.
  - `Query(text)` computes a signature, collects candidate IDs from band buckets, and scores them via Jaccard similarity over MinHash signatures.
  - `FullScanDuplicates(threshold)` runs LSH-based candidate generation over all documents and returns unique pairs above a given similarity threshold.

## Functional Testing

All three components include randomized or scenario-based tests using `go test`:

- `**internal/hashfs`**:
  - Inserts thousands of randomly generated key/value pairs, verifies `Get` against a reference `map[string][]byte`.
  - Performs random updates and deletes, then re-validates against the map.
  - Checks persistence by reopening the on-disk store and reading previously written keys.
- `**internal/perfecthash**`:
  - Builds a table from thousands of unique random byte keys and asserts that every key is found and index is in range.
  - Mutated keys (one-byte difference) are rarely found, validating low false positive rate.
  - Serialization/deserialization round-trip preserves lookup behavior.
- `**internal/lshtext**`:
  - Builds an index over a small set of short English sentences with near-duplicates and unrelated noise.
  - Ensures that a near-duplicate document is returned as a candidate with similarity above 0.5.
  - `FullScanDuplicates` finds at least one pair of near-duplicates.

All tests currently pass: `go test ./internal/...`.

## Performance Measurements

Benchmarks were written as `*_bench_test.go` tests and run via `go test -bench`.
Below are representative results from one run on the given Linux machine.

### HashFS (`internal/hashfs`)

Benchmarks (bucket count `1<<20`, small fixed-size keys/values):

- **Bulk insert** — `BenchmarkStore_Insert`:
  - ~~1.0 µs/op (~~1.0M inserts/sec).
- **Random get** — `BenchmarkStore_RandomGet`:
  - ~1.2 µs/op.
- **Update-heavy** — `BenchmarkStore_UpdateHeavy`:
  - ~0.9 µs/op; similar cost to insert due to append-only log writes.
- **File growth** — `BenchmarkStore_FileSize`:
  - After 100k inserts, file size ≈ 11.8 MB, consistent with record overhead (headers + next pointers).

### PerfectHash (`internal/perfecthash`)

- **Build** — `BenchmarkPerfectHashBuild` (100k keys):
  - ~9.8–10.4 ms per full build; dominated by map allocations and string conversions.
- **Lookup** — `BenchmarkPerfectHashLookup`:
  - ~21 ns/op, effectively O(1) with very low constant.

### LSH for texts (`internal/lshtext`)

- **Index build** — `BenchmarkLSHIndexBuild` (50k short sentences reused cyclically):
  - ~10–11 µs/op for adding a document (including shingling, MinHash, and LSH band updates).
- **Query** — `BenchmarkLSHQuery` on a typical sentence:
  - ~0.5 ms/op, reflecting cost of MinHash computation and scanning LSH buckets.

These numbers meet the homework’s spirit: all operations are fast enough to scale to ≈1M elements with reasonable throughput.

## Profiling (CPU & Memory)

CPU and memory profiles were collected for a representative benchmark of each component using:

- `go test -bench=BenchmarkStore_Insert -cpuprofile=hashfs_cpu.out -memprofile=hashfs_mem.out ./internal/hashfs`
- `go test -bench=BenchmarkPerfectHashBuild -cpuprofile=perfecthash_cpu.out -memprofile=perfecthash_mem.out ./internal/perfecthash`
- `go test -bench=BenchmarkLSHIndexBuild -cpuprofile=lsh_cpu.out -memprofile=lsh_mem.out ./internal/lshtext`

Below — качественные выводы по «горячим местам» и возможным улучшениям.

### HashFS

- **CPU hotspots** (по ожидаемому анализу `pprof`):
  - FNV-1a хеширование ключей при `Put` и `Get`.
  - Системные вызовы `ReadAt`/`WriteAt` при обходе цепочек бакетов и записи новых записей.
  - Копирование ключей и значений в/из временных буферов.
- **Память**:
  - Основной объём расходуется на `mmap` header+bucket-таблицы и на рост файла данных; куча используется умеренно (буферы в `appendRecord`/`readRecord`).
- **Бутылочные горлышки и гипотезы оптимизаций**:
  - **Хеш-функция**: заменить FNV-1a на более быстрый хеш, особенно для небольших ключей (например, собственная не-криптографическая функция над `uint64`/4-байтовыми ключами), чтобы снизить долю времени на CPU.
  - **I/O батчинг**: уменьшить количество отдельных `ReadAt`/`WriteAt` для последовательных записей (буферизация серии вставок, предвыделение страниц данных).
  - **Компактизация**: реализовать периодический compaction, чтобы уменьшать длину цепочек, объём tombstone-записей и размер файла, что снизит как нагрузку на диск, так и количество чтений.

### PerfectHash

- **CPU hotspots**:
  - Построение `map[string]int` — аллокации и хеширование строк при `Build`.
  - Конвертация `[]byte` → `string` в `Build` и `Lookup`.
- **Память**:
  - Основные аллокации — для ключей в `map` и на слайс результата при сериализации.
- **Бутылочные горлышки и гипотезы оптимизаций**:
  - **Избежать копий ключей**: использовать `string`-представление, построенное один раз с переиспользованием, или перейти на собственную хеш-таблицу поверх `[]byte`, чтобы не создавать дополнительные строки.
  - **Более компактная структура**: реализовать настоящий двухуровневый perfect hash (как в исходном плане) с предвычисленными параметрами и фиксированным массивом; это уменьшит аллокации и улучшит предсказуемость кеша.

### LSH for texts

- **CPU hotspots**:
  - Токенизация (`strings.Fields`) и шинглинг текста.
  - MinHash: многократные вызовы FNV-1a для каждой комбинации шингл × хеш-функция.
  - Хеширование band-подписей (сборка `hashBand`) при добавлении и запросах.
- **Память**:
  - Подписи документа (`[]uint64` на `sigSize` элементов) + карты bucket-ов; при большом количестве документов это основное хранилище.
- **Бутылочные горлышки и гипотезы оптимизаций**:
  - **Снижение `sigSize`**: уменьшение длины подписи (например, до 32) сократит как CPU, так и память с небольшим ростом вероятности коллизий.
  - **Кэширование шинглов/подписей**: для часто похожих или идентичных документов можно переиспользовать подписанные представления.
  - **Оптимизация хеш-функций**: использовать более простую (но всё ещё равномерную) integer-хеш-функцию вместо FNV-1a для внутренних шагов MinHash/LSH.

## How to Reproduce

- Запустить все тесты:
  ```bash

  ```

go test ./internal/...

- Собрать профили и изучить их в `pprof` (пример для `hashfs`):
  ```bash

  ```

go test -bench=BenchmarkStore_Insert -cpuprofile=hashfs_cpu.out -memprofile=hashfs_mem.out ./internal/hashfs
go tool pprof ./internal/hashfs.test hashfs_cpu.out