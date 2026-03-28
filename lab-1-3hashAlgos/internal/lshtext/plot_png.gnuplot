set datafile separator ","
set terminal png size 900,550 enhanced font "Arial,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

set output plot_dir . "/benchmark_latency.png"
set title "Latency per document — LSH index"
set xlabel "Corpus size (N)"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "build", \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):5 \
      with lines lw 1 lc rgb "#e41a1c" notitle, \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars lw 2 pt 9 lc rgb "#377eb8" title "add", \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):5 \
      with lines lw 1 lc rgb "#377eb8" notitle, \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars lw 2 pt 5 lc rgb "#4daf4a" title "full scan", \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):5 \
      with lines lw 1 lc rgb "#4daf4a" notitle

set output plot_dir . "/benchmark_throughput.png"
set title "Throughput — LSH index"
set xlabel "Corpus size (N)"
set ylabel "docs/s"
plot \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):7 \
      with linespoints lw 2 pt 7 lc rgb "#e41a1c" title "build", \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):7 \
      with linespoints lw 2 pt 9 lc rgb "#377eb8" title "add", \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):7 \
      with linespoints lw 2 pt 5 lc rgb "#4daf4a" title "full scan"
