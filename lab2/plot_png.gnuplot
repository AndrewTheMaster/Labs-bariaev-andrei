set datafile separator ","
set terminal png size 960,580 enhanced font "Arial,11"
set key outside right top
set grid
set border linewidth 1.2
set pointsize 1.2
set logscale x 10
set autoscale xfix

bench_csv = raw_dir . "/benchmarks.csv"

# ── Insert latency ────────────────────────────────────────────────────────────
set output plot_dir . "/insert_latency.png"
set title "Insert latency per point"
set xlabel "Dataset size (N)"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "geohash-p5", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#e41a1c" notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p6" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 9 lc rgb "#377eb8" title "geohash-p6", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "geohash-p6" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#377eb8" notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "kdtree" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 5 lc rgb "#4daf4a" title "kd-tree", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "kdtree" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#4daf4a" notitle, \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "brute" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 11 lc rgb "#984ea3" title "brute", \
  bench_csv using (strcol(1) eq "Insert" && strcol(2) eq "brute" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#984ea3" notitle

# ── FindNearby latency ────────────────────────────────────────────────────────
set output plot_dir . "/find_latency.png"
set title "FindNearby latency (r = 10 km)"
set xlabel "Dataset size (N)"
set ylabel "ns/query"
plot \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "geohash-p5", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "geohash-p5" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#e41a1c" notitle, \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-balanced" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 5 lc rgb "#4daf4a" title "kd-tree (balanced)", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-balanced" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#4daf4a" notitle, \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-online" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 9 lc rgb "#ff7f00" title "kd-tree (online)", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "kdtree-online" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#ff7f00" notitle, \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "brute" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 11 lc rgb "#984ea3" title "brute", \
  bench_csv using (strcol(1) eq "FindNearby_VaryN" && strcol(2) eq "brute" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#984ea3" notitle

# ── KD-Tree build ─────────────────────────────────────────────────────────────
set output plot_dir . "/kdtree_build.png"
set title "KD-Tree build time per point"
set xlabel "Dataset size (N)"
set ylabel "ns/item"
plot \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "online" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 7 lc rgb "#e41a1c" title "online insert", \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "online" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#e41a1c" notitle, \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "balanced" ? $3 : 1/0):6:($6-$7):($6+$7) \
      with yerrorbars lw 2 pt 9 lc rgb "#377eb8" title "balanced build", \
  bench_csv using (strcol(1) eq "BuildKDTree" && strcol(2) eq "balanced" ? $3 : 1/0):6 \
      with lines lw 1 lc rgb "#377eb8" notitle
