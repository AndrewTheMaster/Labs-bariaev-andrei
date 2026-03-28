set datafile separator ","
set terminal pdfcairo enhanced color size 8in,5in font "Helvetica,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

# ── Задержка с доверительными интервалами ───────────────────────────────────
set output plot_dir . "/benchmark_latency.pdf"
set title "latency per item"
set xlabel "Corpus size"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars linewidth 2 pt 7 title "build", \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):5 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars linewidth 2 pt 9 title "add", \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):5 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):5:($5-$6):($5+$6) \
      with yerrorbars linewidth 2 pt 5 title "full scan", \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):5 \
      with lines linewidth 1 notitle

# ── Пропускная способность ───────────────────────────────────────────────────
set output plot_dir . "/benchmark_throughput.pdf"
set title "throughput"
set xlabel "Corpus size"
set ylabel "ops/s"
plot \
  bench_csv using (strcol(1) eq "build" ? $2 : 1/0):7 \
      with linespoints linewidth 2 pt 7 title "build", \
  bench_csv using (strcol(1) eq "add" ? $2 : 1/0):7 \
      with linespoints linewidth 2 pt 9 title "add", \
  bench_csv using (strcol(1) eq "fullscan" ? $2 : 1/0):7 \
      with linespoints linewidth 2 pt 5 title "full scan"
