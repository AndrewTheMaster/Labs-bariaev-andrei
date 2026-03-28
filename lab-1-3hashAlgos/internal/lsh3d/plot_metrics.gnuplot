set datafile separator ","
set terminal pdfcairo size 16cm,10cm enhanced font "Arial,10"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

filter(b,i) = (strcol(1) eq b) ? column(i) : 1/0

set output plot_dir . "/benchmark_latency.pdf"
set title "Latency per point — LSH 3D index"
set xlabel "Dataset size (N)"
set ylabel "ns/item"
plot \
  bench_csv using (filter("build",2)):(filter("build",5)):(filter("build",5)-filter("build",6)):(filter("build",5)+filter("build",6)) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "build", \
  bench_csv using (filter("build",2)):(filter("build",5)) \
      with lines lw 1 lc rgb "#e41a1c" notitle, \
  bench_csv using (filter("add",2)):(filter("add",5)):(filter("add",5)-filter("add",6)):(filter("add",5)+filter("add",6)) \
      with yerrorbars lw 2 pt 9 lc rgb "#377eb8" title "add", \
  bench_csv using (filter("add",2)):(filter("add",5)) \
      with lines lw 1 lc rgb "#377eb8" notitle, \
  bench_csv using (filter("query",2)):(filter("query",5)):(filter("query",5)-filter("query",6)):(filter("query",5)+filter("query",6)) \
      with yerrorbars lw 2 pt 5 lc rgb "#4daf4a" title "query", \
  bench_csv using (filter("query",2)):(filter("query",5)) \
      with lines lw 1 lc rgb "#4daf4a" notitle, \
  bench_csv using (filter("fullscan",2)):(filter("fullscan",5)):(filter("fullscan",5)-filter("fullscan",6)):(filter("fullscan",5)+filter("fullscan",6)) \
      with yerrorbars lw 2 pt 13 lc rgb "#ff7f00" title "full scan", \
  bench_csv using (filter("fullscan",2)):(filter("fullscan",5)) \
      with lines lw 1 lc rgb "#ff7f00" notitle

set output plot_dir . "/benchmark_throughput.pdf"
set title "Throughput — LSH 3D index"
set xlabel "Dataset size (N)"
set ylabel "points/s"
plot \
  bench_csv using (filter("build",2)):(filter("build",7)) \
      with linespoints lw 2 pt 7 lc rgb "#e41a1c" title "build", \
  bench_csv using (filter("add",2)):(filter("add",7)) \
      with linespoints lw 2 pt 9 lc rgb "#377eb8" title "add", \
  bench_csv using (filter("query",2)):(filter("query",7)) \
      with linespoints lw 2 pt 5 lc rgb "#4daf4a" title "query", \
  bench_csv using (filter("fullscan",2)):(filter("fullscan",7)) \
      with linespoints lw 2 pt 13 lc rgb "#ff7f00" title "full scan"
