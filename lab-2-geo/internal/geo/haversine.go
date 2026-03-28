// Package geo содержит геометрические примитивы: формулу Хаверсина и геохэш.
package geo

import "math"

const earthRadiusKm = 6371.0

// DistanceKm вычисляет расстояние между двумя точками (lat1,lng1) и (lat2,lng2)
// по формуле Хаверсина. Возвращает результат в километрах.
//
// Формула:
//
//	a = sin²(Δlat/2) + cos(lat1)·cos(lat2)·sin²(Δlng/2)
//	d = 2·R·arcsin(√a)
func DistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	lat1Rad := toRad(lat1)
	lat2Rad := toRad(lat2)

	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	a := sinDLat*sinDLat + math.Cos(lat1Rad)*math.Cos(lat2Rad)*sinDLng*sinDLng
	c := 2 * math.Asin(math.Sqrt(a))
	return earthRadiusKm * c
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}
