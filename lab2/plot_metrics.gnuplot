set datafile separator ","
set terminal pdfcairo enhanced color size 8in,5in font "Helvetica,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

filter(b,i) = "< awk -F',' 'NR>1 && $1==\"" . b . "\" && $2==\"" . i . "\"' " . bench_csv

# ── Insert: задержка с доверительными интервалами ─────────────────────────────
set output plot_dir . "/insert_latency.pdf"
set title "Insert latency per point"
set xlabel "Dataset size"
set ylabel "ns/item"
unset logscale y
plot \
  filter("Insert","geohash-p5") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 title "geohash-p5", \
  filter("Insert","geohash-p6") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 title "geohash-p6", \
  filter("Insert","kdtree") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 5 title "kd-tree", \
  filter("Insert","brute") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 11 title "brute"

# ── Insert: пропускная способность ───────────────────────────────────────────
set output plot_dir . "/insert_throughput.pdf"
set title "Insert throughput"
set xlabel "Dataset size"
set ylabel "ops/s"
plot \
  filter("Insert","geohash-p5") using 3:8 with linespoints lw 2 pt 7 title "geohash-p5", \
  filter("Insert","geohash-p6") using 3:8 with linespoints lw 2 pt 9 title "geohash-p6", \
  filter("Insert","kdtree")     using 3:8 with linespoints lw 2 pt 5 title "kd-tree", \
  filter("Insert","brute")      using 3:8 with linespoints lw 2 pt 11 title "brute"

# ── FindNearby: задержка (log Y — разброс ~100x между brute и kd-tree) ────────
set output plot_dir . "/find_latency.pdf"
set title "FindNearby latency (r=10 km)"
set xlabel "Dataset size"
set ylabel "ns/query"
set logscale y 10
plot \
  filter("FindNearby_VaryN","geohash-p5") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 title "geohash-p5", \
  filter("FindNearby_VaryN","kdtree-balanced") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 5 title "kd-tree (balanced)", \
  filter("FindNearby_VaryN","kdtree-online") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 title "kd-tree (online)", \
  filter("FindNearby_VaryN","brute") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 11 title "brute"

# ── KD-Tree build: online vs balanced ────────────────────────────────────────
set output plot_dir . "/kdtree_build.pdf"
set title "KD-Tree build time per point"
set xlabel "Dataset size"
set ylabel "ns/item"
unset logscale y
plot \
  filter("BuildKDTree","online") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 title "online insert", \
  filter("BuildKDTree","balanced") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 title "balanced build"
