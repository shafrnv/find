package domain

import "fmt"

// GeoPoint — точка на карте. Широта и долгота задаются парой.
type GeoPoint struct {
	Lat float64
	Lon float64
}

func NewGeoPoint(lat, lon float64) (GeoPoint, error) {
	if lat < -90 || lat > 90 {
		return GeoPoint{}, fmt.Errorf("lat: широта от -90 до 90")
	}
	if lon < -180 || lon > 180 {
		return GeoPoint{}, fmt.Errorf("lon: долгота от -180 до 180")
	}
	return GeoPoint{Lat: lat, Lon: lon}, nil
}
