package tracklog

import (
	"fmt"
	"math"
)

type Duration uint // seconds

func (d Duration) String() string {
	seconds := d % 60
	minutes := int(d/60) % 60
	hours := int(d / 60 / 60)

	if hours < 24 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}

	days := int(hours / 24)
	hours %= 24
	if days == 1 {
		return fmt.Sprintf("1 day %d:%02d:%02d", hours, minutes, seconds)
	}

	return fmt.Sprintf("%d days %d:%02d:%02d", days, hours, minutes, seconds)
}

type Distance uint // meters

func (d Distance) String() string {
	km := float64(d) / 1000.0
	return fmt.Sprintf("%.02f km", km)
}

type Speed float64 // meters per second

func (s Speed) String() string {
	kmh := s / 1000.0 * 3600.0
	return fmt.Sprintf("%.02f km/h", kmh)
}

func (s Speed) Pace() Pace {
	return Pace(1 / s)
}

type Pace float64 // seconds per meter

func (p Pace) String() string {
	secPerKM := p * 1000
	min := int(secPerKM / 60)
	sec := int(secPerKM) % 60
	return fmt.Sprintf("%d:%02d min/km", min, sec)
}

const earthRadius = 6371000

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	rlat1 := lat1 * (math.Pi / 180.0)
	rlat2 := lat2 * (math.Pi / 180.0)
	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(rlat1) * math.Cos(rlat2)
	a := a1 + a2
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}
