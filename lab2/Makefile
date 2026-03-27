SIZES        ?= 1000,10000,50000,100000,500000
METRICS_OUT  ?= metrics
FUZZ_TIME    ?= 15s
BENCH_FILTER ?= ^Benchmark
RAW_DIR      := $(METRICS_OUT)/raw
PLOT_DIR     := $(METRICS_OUT)/plots

.PHONY: test test-short fuzz bench collect metrics plot clean

test:
	go test ./...

test-short:
	go test -short ./...

fuzz:
	go test -run '$$^' -fuzz '^FuzzSearcherAgainstBrute$$' -fuzztime $(FUZZ_TIME) ./internal/geoindex/

bench:
	SIZES='$(SIZES)' go test -run '$$^' -bench '$(BENCH_FILTER)' -benchmem ./internal/geoindex/

collect:
	mkdir -p $(RAW_DIR) $(PLOT_DIR)
	SIZES='$(SIZES)' go test -run '$$^' -bench '$(BENCH_FILTER)' -benchmem ./internal/geoindex/ > $(RAW_DIR)/benchmarks.txt
	awk 'BEGIN { OFS=","; print "operation,impl,size,iterations,ns_per_op,ns_per_item,ci95_ns_per_item,ops_per_sec,bytes_per_op" } \
	/^Benchmark(Insert|FindNearby_VaryN|BuildKDTree)\// { \
	  name=$$1; iters=$$2; \
	  split(name, a, "/"); bench=a[1]; sub("^Benchmark", "", bench); \
	  impl=""; size=0; \
	  for (k=2; k<=length(a); k++) { \
	    if (a[k] ~ /^size=/) { split(a[k], s, "="); split(s[2], sd, "-"); size=sd[1]; } \
	    else { split(a[k], sd, "-"); impl=sd[1]; } \
	  } \
	  delete m; \
	  for (i=3; i+1<=NF; i+=2) m[$$(i+1)]=$$i; \
	  ns_item=("ns/item" in m)?m["ns/item"]:0; \
	  ci=("ci95_ns/item" in m)?m["ci95_ns/item"]:0; \
	  ops=("ops/s" in m)?m["ops/s"]:0; \
	  bop=("B/op" in m)?m["B/op"]:0; \
	  nsop=("ns/op" in m)?m["ns/op"]:0; \
	  print bench,impl,size,iters,nsop,ns_item,ci,ops,bop; \
	}' $(RAW_DIR)/benchmarks.txt > $(RAW_DIR)/benchmarks.csv

metrics: collect plot

plot:
	mkdir -p $(PLOT_DIR)
	gnuplot -e "raw_dir='$(RAW_DIR)'; plot_dir='$(PLOT_DIR)'" plot_metrics.gnuplot

clean:
	rm -rf metrics
