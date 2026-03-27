// Package brute предоставляет эталонный алгоритм геопоиска — полный перебор.
//
// Полный перебор (linear scan) проверяет расстояние до каждой точки в индексе.
// Сложность:
//   - Insert: O(1) (append)
//   - FindNearby: O(N) — всегда просматривает весь набор данных
//
// Используется как baseline в бенчмарках для демонстрации выигрыша
// геохэш-индекса и k-d дерева при больших N.
package brute

import (
	"sort"

	"siaod-hw2/internal/geo"
	"siaod-hw2/internal/geoindex"
)

// Scanner — реализация geoindex.Searcher методом полного перебора.
type Scanner struct {
	points []geoindex.Point
}

// New создаёт пустой сканер.
func New() *Scanner { return &Scanner{} }

// Insert добавляет точку (O(1) amortized).
func (s *Scanner) Insert(p geoindex.Point) {
	s.points = append(s.points, p)
}

// FindNearby возвращает все точки в радиусе radiusKm (O(N)).
func (s *Scanner) FindNearby(lat, lng, radiusKm float64) []geoindex.Result {
	var results []geoindex.Result
	for i := range s.points {
		d := geo.DistanceKm(lat, lng, s.points[i].Lat, s.points[i].Lng)
		if d <= radiusKm {
			results = append(results, geoindex.Result{Point: s.points[i], Distance: d})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	return results
}

// FindKNearest возвращает k ближайших точек (O(N log k)).
func (s *Scanner) FindKNearest(lat, lng float64, k int) []geoindex.Result {
	results := make([]geoindex.Result, len(s.points))
	for i := range s.points {
		results[i] = geoindex.Result{
			Point:    s.points[i],
			Distance: geo.DistanceKm(lat, lng, s.points[i].Lat, s.points[i].Lng),
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	if k < len(results) {
		return results[:k]
	}
	return results
}

// Count возвращает число точек.
func (s *Scanner) Count() int { return len(s.points) }
