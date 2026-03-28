set datafile separator ","
set terminal png size 960,580 enhanced font "Arial,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

# helper macros: filter CSV rows by benchmark name and impl tag
# usage: plot filter("Insert","geohash-p5") using 3:6:($6-$7):($6+$7) with yerrorlines ...
filter(b,i) = "< awk -F',' 'NR>1 && $1==\"" . b . "\" && $2==\"" . i . "\"' " . bench_csv

# ── Insert latency ────────────────────────────────────────────────────────────
set output plot_dir . "/insert_latency.png"
set title "Insert latency per point"
set xlabel "Dataset size (N)"
set ylabel "ns/item"
unset logscale y
plot \
  filter("Insert","geohash-p5") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 lc rgb "#e41a1c" title "geohash-p5", \
  filter("Insert","geohash-p6") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 lc rgb "#377eb8" title "geohash-p6", \
  filter("Insert","kdtree") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 5 lc rgb "#4daf4a" title "kd-tree", \
  filter("Insert","brute") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 11 lc rgb "#984ea3" title "brute"

# ── FindNearby latency (log Y — разброс ~100x между brute и kd-tree) ─────────
set output plot_dir . "/find_latency.png"
set title "FindNearby latency (r = 10 km)"
set xlabel "Dataset size (N)"
set ylabel "ns/query"
set logscale y 10
plot \
  filter("FindNearby_VaryN","geohash-p5") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 lc rgb "#e41a1c" title "geohash-p5", \
  filter("FindNearby_VaryN","kdtree-balanced") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 5 lc rgb "#4daf4a" title "kd-tree (balanced)", \
  filter("FindNearby_VaryN","kdtree-online") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 lc rgb "#ff7f00" title "kd-tree (online)", \
  filter("FindNearby_VaryN","brute") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 11 lc rgb "#984ea3" title "brute"

# ── KD-Tree build ─────────────────────────────────────────────────────────────
set output plot_dir . "/kdtree_build.png"
set title "KD-Tree build time per point"
set xlabel "Dataset size (N)"
set ylabel "ns/item"
unset logscale y
plot \
  filter("BuildKDTree","online") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 7 lc rgb "#e41a1c" title "online insert", \
  filter("BuildKDTree","balanced") using 3:6:($6-$7):($6+$7) \
      with yerrorlines lw 2 pt 9 lc rgb "#377eb8" title "balanced build"
