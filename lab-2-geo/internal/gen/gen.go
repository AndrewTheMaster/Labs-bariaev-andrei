// Package gen генерирует реалистичные наборы географических данных для тестов и бенчмарков.
package gen

import (
	"fmt"
	mathrand "math/rand"

	"siaod-hw2/internal/geoindex"
)

// NewRand создаёт детерминированный генератор с заданным seed.
func NewRand(seed int64) *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(seed))
}

// Регионы, имитирующие реальные скопления POI (Points of Interest).
// Каждый регион задаётся центром и радиусом разброса (в градусах).
var regions = []struct {
	name       string
	lat, lng   float64
	spreadLat  float64
	spreadLng  float64
}{
	{"Moscow",     55.75, 37.62, 0.5, 0.8},
	{"Saint-Petersburg", 59.94, 30.32, 0.4, 0.6},
	{"Novosibirsk", 54.99, 82.90, 0.3, 0.5},
	{"Yekaterinburg", 56.84, 60.61, 0.3, 0.5},
	{"London",     51.51, -0.13, 0.4, 0.7},
	{"Paris",      48.85, 2.35, 0.3, 0.5},
	{"Berlin",     52.52, 13.40, 0.3, 0.5},
	{"New York",   40.71, -74.01, 0.4, 0.6},
	{"Tokyo",      35.68, 139.69, 0.4, 0.6},
	{"Beijing",    39.91, 116.39, 0.4, 0.6},
}

// RandomRegionPoint генерирует точку в одном из заданных регионов.
// 80% точек попадают в регионы (имитируя скопления), 20% — равномерно по миру.
func RandomRegionPoint(r *mathrand.Rand, id int) geoindex.Point {
	var lat, lng float64
	if r.Float64() < 0.8 {
		reg := regions[r.Intn(len(regions))]
		lat = reg.lat + (r.Float64()*2-1)*reg.spreadLat
		lng = reg.lng + (r.Float64()*2-1)*reg.spreadLng
	} else {
		// равномерно по всей Земле
		lat = r.Float64()*170 - 85
		lng = r.Float64()*360 - 180
	}
	return geoindex.Point{
		ID:   fmt.Sprintf("poi-%08d", id),
		Lat:  lat,
		Lng:  lng,
		Data: []byte(fmt.Sprintf(`{"name":"place-%d","type":"poi"}`, id)),
	}
}

// RandomUniformPoint генерирует равномерно распределённую точку по всей Земле.
func RandomUniformPoint(r *mathrand.Rand, id int) geoindex.Point {
	lat := r.Float64()*170 - 85
	lng := r.Float64()*360 - 180
	return geoindex.Point{
		ID:  fmt.Sprintf("poi-%08d", id),
		Lat: lat,
		Lng: lng,
	}
}

// GeneratePoints генерирует n точек, используя RandomRegionPoint.
func GeneratePoints(r *mathrand.Rand, n int) []geoindex.Point {
	pts := make([]geoindex.Point, n)
	for i := range pts {
		pts[i] = RandomRegionPoint(r, i)
	}
	return pts
}

// NearbyPoint генерирует точку в радиусе ~radiusKm от (lat, lng).
// Приближённый перевод: 1° широты ≈ 111 km.
func NearbyPoint(r *mathrand.Rand, lat, lng, radiusKm float64, id int) geoindex.Point {
	dLat := (r.Float64()*2 - 1) * radiusKm / 111.0
	dLng := (r.Float64()*2 - 1) * radiusKm / 111.0
	return geoindex.Point{
		ID:  fmt.Sprintf("near-%08d", id),
		Lat: lat + dLat,
		Lng: lng + dLng,
	}
}
