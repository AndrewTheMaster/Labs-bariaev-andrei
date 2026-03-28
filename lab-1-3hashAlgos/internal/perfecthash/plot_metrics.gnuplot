set datafile separator ","
set terminal pdfcairo enhanced color size 8in,5in font "Helvetica,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

# ── Средняя задержка с доверительными интервалами ───────────────────────────
set output plot_dir . "/benchmark_latency.pdf"
set title "average time per one operation"
set xlabel "rows"
set ylabel "ns/op"
plot \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars linewidth 2 pt 7 title "get_avg_ns", \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):5 \
      with lines linewidth 1 notitle

# ── Пропускная способность ───────────────────────────────────────────────────
set output plot_dir . "/benchmark_throughput.pdf"
set title "throughput per one element"
set xlabel "rows"
set ylabel "ops/sec"
plot \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):7 \
      with linespoints linewidth 2 pt 7 title "get_ops_sec"
