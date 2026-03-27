// Package geoindex реализует геопространственный индекс на основе геохэша.
//
// Алгоритм:
//  1. При вставке объекта его координаты кодируются в геохэш заданной точности.
//  2. Объект помещается в bucket[geohash].
//  3. При поиске вычисляется геохэш запросной точки, затем берутся сама ячейка
//     и её 8 соседей (3×3 сетка). Из всех найденных кандидатов оставляются
//     только те, что попадают в радиус запроса по формуле Хаверсина.
//
// Сложность:
//   - Insert: O(1) (hash + map lookup)
//   - FindNearby: O(k), где k — среднее число точек в 9 ячейках
package geoindex

import (
	"sort"

	"siaod-hw2/internal/geo"
)

// Point — географическая точка с произвольными метаданными.
type Point struct {
	ID   string
	Lat  float64
	Lng  float64
	Data []byte
}

// Result — результат поиска: точка + расстояние до запросной координаты.
type Result struct {
	Point    Point
	Distance float64 // км
}

// Searcher — общий интерфейс для всех реализаций геопоиска.
// Используется в бенчмарках для сравнения алгоритмов.
type Searcher interface {
	// Insert добавляет точку в индекс.
	Insert(p Point)
	// FindNearby возвращает все точки в радиусе radiusKm вокруг (lat,lng),
	// отсортированные по расстоянию.
	FindNearby(lat, lng, radiusKm float64) []Result
	// Count возвращает число точек в индексе.
	Count() int
}

// Index — геохэш-индекс.
type Index struct {
	precision int
	cells     map[string][]Point
	count     int
}

// New создаёт новый индекс.
//
// precision задаёт длину геохэша (1–12). Рекомендуемые значения:
//   - 4 (~39 km×20 km ячейки): для радиусов 50–200 km
//   - 5 (~4.9 km×4.9 km):      для радиусов 5–50 km
//   - 6 (~1.2 km×0.6 km):      для радиусов 1–5 km
func New(precision int) *Index {
	if precision < 1 {
		precision = 1
	}
	if precision > 12 {
		precision = 12
	}
	return &Index{
		precision: precision,
		cells:     make(map[string][]Point),
	}
}

// Insert добавляет точку в индекс.
func (idx *Index) Insert(p Point) {
	h := geo.Encode(p.Lat, p.Lng, idx.precision)
	idx.cells[h] = append(idx.cells[h], p)
	idx.count++
}

// FindNearby возвращает все точки в радиусе radiusKm вокруг (lat, lng).
// Результаты отсортированы по расстоянию (ближайшие первые).
//
// Алгоритм выборки ячеек:
//  1. Строим охватывающий прямоугольник круга (bounding box).
//  2. Сканируем прямоугольник с шагом ≈ 0.5 высоты/ширины ячейки,
//     кодируя каждую точку в геохэш → получаем множество уникальных ячеек.
//  3. Для каждой кандидатной ячейки берём её точки и фильтруем по Хаверсину.
//
// Сложность: O(k * m) где k — число ячеек в bounding box, m — среднее заполнение ячейки.
func (idx *Index) FindNearby(lat, lng, radiusKm float64) []Result {
	cells := idx.candidateCells(lat, lng, radiusKm)

	var results []Result
	for _, cell := range cells {
		for _, p := range idx.cells[cell] {
			d := geo.DistanceKm(lat, lng, p.Lat, p.Lng)
			if d <= radiusKm {
				results = append(results, Result{Point: p, Distance: d})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	return results
}

// candidateCells возвращает все геохэш-ячейки, перекрывающиеся с кругом (lat, lng, radiusKm).
//
// Алгоритм: BFS-расширение из центральной ячейки через соседей.
// Останавливаемся, когда центр очередной ячейки дальше, чем radius + диагональ ячейки.
// Сложность: O(k), где k — число ячеек, покрывающих круг.
func (idx *Index) candidateCells(lat, lng, radiusKm float64) []string {
	// Размер ячейки для оценки порога включения.
	_, _, latErr, lngErr, _ := geo.Decode(geo.Encode(lat, lng, idx.precision))
	// Диагональ ячейки в км (с запасом).
	cellDiag := geo.DistanceKm(
		lat-latErr, lng-lngErr,
		lat+latErr, lng+lngErr,
	)
	cutoff := radiusKm + cellDiag

	startHash := geo.Encode(lat, lng, idx.precision)

	visited := make(map[string]struct{})
	queue := []string{startHash}
	visited[startHash] = struct{}{}
	var result []string

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		// Проверяем, что центр ячейки в пределах cutoff.
		cLat, cLng, _, _, _ := geo.Decode(cur)
		if geo.DistanceKm(lat, lng, cLat, cLng) > cutoff {
			continue
		}
		result = append(result, cur)

		for _, nb := range geo.Neighbors(cur) {
			if _, ok := visited[nb]; !ok {
				visited[nb] = struct{}{}
				queue = append(queue, nb)
			}
		}
	}

	return result
}

// FindKNearest возвращает k ближайших точек. Если точек в индексе меньше k,
// возвращает все доступные. Использует адаптивное расширение радиуса поиска.
func (idx *Index) FindKNearest(lat, lng float64, k int) []Result {
	if k <= 0 || idx.count == 0 {
		return nil
	}

	// Ячейки индекса покрывают фиксированный радиус; мы начинаем с одной ячейки
	// и постепенно увеличиваем охват, пока не наберём k кандидатов или не
	// обойдём весь индекс.
	h := geo.Encode(lat, lng, idx.precision)
	visited := make(map[string]struct{})
	frontier := []string{h}
	var all []Result

	for len(frontier) > 0 && len(all) < k {
		nextFrontier := make(map[string]struct{})
		for _, cell := range frontier {
			if _, ok := visited[cell]; ok {
				continue
			}
			visited[cell] = struct{}{}
			for _, p := range idx.cells[cell] {
				d := geo.DistanceKm(lat, lng, p.Lat, p.Lng)
				all = append(all, Result{Point: p, Distance: d})
			}
			for _, nb := range geo.Neighbors(cell) {
				if _, ok := visited[nb]; !ok {
					nextFrontier[nb] = struct{}{}
				}
			}
		}
		frontier = frontier[:0]
		for nb := range nextFrontier {
			frontier = append(frontier, nb)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Distance < all[j].Distance
	})
	if len(all) > k {
		return all[:k]
	}
	return all
}

// Count возвращает число точек в индексе.
func (idx *Index) Count() int { return idx.count }

// Precision возвращает текущую точность геохэша.
func (idx *Index) Precision() int { return idx.precision }
