set datafile separator ","
set terminal png size 900,550 enhanced font "Arial,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

set output plot_dir . "/benchmark_latency.png"
set title "Lookup latency — perfect hash"
set xlabel "N (keys)"
set ylabel "ns/op"
plot \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "lookup ± CI95", \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):5 \
      with lines lw 1 lc rgb "#e41a1c" notitle

set output plot_dir . "/benchmark_throughput.png"
set title "Lookup throughput — perfect hash"
set xlabel "N (keys)"
set ylabel "ops/s"
plot \
  bench_csv using (strcol(1) eq "lookup" ? $2 : 1/0):7 \
      with linespoints lw 2 pt 7 lc rgb "#377eb8" title "lookup"
