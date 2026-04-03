SIZES        ?= 1000,5000,10000,50000,100000
METRICS_OUT  ?= metrics
FUZZ_TIME    ?= 15s
BENCH_FILTER ?= ^BenchmarkLSH3D
BENCH_TIME   ?= 20x
RAW_DIR      := $(METRICS_OUT)/raw
PLOT_DIR     := $(METRICS_OUT)/plots
PROF_DIR     := $(METRICS_OUT)/profiles

.PHONY: test fuzz bench collect metrics plot profile clean

test:
	go test ./...

fuzz:
	go test -run '$$^' -fuzz '^FuzzIndexNoPanic$$' -fuzztime $(FUZZ_TIME) .

bench:
	SIZES='$(SIZES)' go test -run '$$^' -bench '$(BENCH_FILTER)' -benchmem -benchtime $(BENCH_TIME) .

collect:
	mkdir -p $(RAW_DIR) $(PLOT_DIR)
	SIZES='$(SIZES)' go test -run '$$^' -bench '$(BENCH_FILTER)' -benchmem -benchtime $(BENCH_TIME) . > $(RAW_DIR)/benchmarks.txt
	awk 'BEGIN { OFS=","; print "operation,size,iterations,ns_per_op,ns_per_item,ci95_ns_per_item,ops_per_sec,bytes_per_op,allocs_per_op" } \
	/^BenchmarkLSH3D(Build|Add|Query|FullScan)\/size=[0-9]+-[0-9]+/ { \
	  name=$$1; iters=$$2; \
	  split(name, a, "/"); split(a[1], b, "BenchmarkLSH3D"); op=tolower(b[2]); \
	  split(a[2], c, "="); split(c[2], d, "-"); size=d[1]; \
	  delete m; \
	  for (i=3; i+1<=NF; i+=2) m[$$(i+1)]=$$i; \
	  ns_item=("ns/item" in m)?m["ns/item"]:0; \
	  ci=("ci95_ns/item" in m)?m["ci95_ns/item"]:0; \
	  ops=("ops/s" in m)?m["ops/s"]:0; \
	  bop=("B/op" in m)?m["B/op"]:0; \
	  alloc=("allocs/op" in m)?m["allocs/op"]:0; \
	  nsop=("ns/op" in m)?m["ns/op"]:0; \
	  print op,size,iters,nsop,ns_item,ci,ops,bop,alloc; \
	}' $(RAW_DIR)/benchmarks.txt > $(RAW_DIR)/benchmarks.csv

metrics: collect plot

plot:
	mkdir -p $(PLOT_DIR)
	gnuplot -e "raw_dir='$(RAW_DIR)'; plot_dir='$(PLOT_DIR)'" plot_png.gnuplot

profile:
	mkdir -p $(PROF_DIR) $(PLOT_DIR)
	SIZES='10000,100000' go test -run '$$^' -bench '^BenchmarkLSH3DBuild$$' -benchmem \
		-benchtime $(BENCH_TIME) \
		-cpuprofile $(PROF_DIR)/cpu_build.prof \
		-memprofile $(PROF_DIR)/mem_build.prof . > /dev/null
	SIZES='10000,100000' go test -run '$$^' -bench '^BenchmarkLSH3DQuery$$' -benchmem \
		-benchtime $(BENCH_TIME) \
		-cpuprofile $(PROF_DIR)/cpu_query.prof \
		-memprofile $(PROF_DIR)/mem_query.prof . > /dev/null
	go tool pprof -top $(PROF_DIR)/cpu_build.prof  > $(PROF_DIR)/cpu_build_top.txt
	go tool pprof -top $(PROF_DIR)/cpu_query.prof  > $(PROF_DIR)/cpu_query_top.txt
	go tool pprof -top $(PROF_DIR)/mem_build.prof  > $(PROF_DIR)/mem_build_top.txt
	go tool pprof -top $(PROF_DIR)/mem_query.prof  > $(PROF_DIR)/mem_query_top.txt
	bash ../../scripts/gen_flamegraphs.sh $(PROF_DIR) $(PLOT_DIR)

clean:
	rm -rf metrics
