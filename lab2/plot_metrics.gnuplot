set datafile separator ","
set terminal pdfcairo enhanced color size 8in,5in font "Helvetica,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

# ── Insert: задержка с доверительными интервалами ─────────────────────────────
set output plot_dir . "/insert_latency.pdf"
set title "Insert latency per point"
set xlabel "Dataset size"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 7 title "geohash-p5", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p6" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 9 title "geohash-p6", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p6" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "kdtree" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 5 title "kdtree", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "kdtree" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "brute" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 11 title "brute", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "brute" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle

# ── Insert: пропускная способность ───────────────────────────────────────────
set output plot_dir . "/insert_throughput.pdf"
set title "Insert throughput"
set xlabel "Dataset size"
set ylabel "ops/s"
plot \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p5" ? $3 : 1/0):8 \
      with linespoints linewidth 2 pt 7 title "geohash-p5", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p6" ? $3 : 1/0):8 \
      with linespoints linewidth 2 pt 9 title "geohash-p6", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "kdtree" ? $3 : 1/0):8 \
      with linespoints linewidth 2 pt 5 title "kdtree", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "brute" ? $3 : 1/0):8 \
      with linespoints linewidth 2 pt 11 title "brute"

# ── FindNearby: задержка с доверительными интервалами ────────────────────────
set output plot_dir . "/find_latency.pdf"
set title "FindNearby latency (r=10 km)"
set xlabel "Dataset size"
set ylabel "ns/query"
plot \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 7 title "geohash-p5", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-balanced" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 5 title "kdtree-balanced", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-balanced" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "brute" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 11 title "brute", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "brute" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle

# ── KD-Tree build: online vs balanced ────────────────────────────────────────
set output plot_dir . "/kdtree_build.pdf"
set title "KD-Tree build time per point"
set xlabel "Dataset size"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "online" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 7 title "online insert", \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "online" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle, \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "balanced" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars linewidth 2 pt 9 title "balanced build", \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "balanced" ? $3 : 1/0):6 \
      with lines linewidth 1 notitle
