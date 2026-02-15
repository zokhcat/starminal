package utils

import (
	"math"
	"time"
)

type VisibleStar struct {
	Alt   float64
	Az    float64
	Mag   float64
	Name  string
	Ci    float64
	Spect string
}

func JulianDate(t time.Time) float64 {
	y := float64(t.Year())
	m := float64(t.Month())
	d := float64(t.Day()) +
		float64(t.Hour())/24.0 +
		float64(t.Minute())/1440.0 +
		float64(t.Second())/86400.0

	if m <= 2 {
		y--
		m += 12
	}

	A := math.Floor(y / 100.0)
	B := 2 - A + math.Floor(A/4.0)

	return math.Floor(365.25*(y+4716)) + math.Floor(30.6001*(m+1)) + d + B - 1524.5
}

func GMST(jd float64) float64 {
	d := jd - 2451545.0
	gmst := 280.46061837 + 360.98564736629*d
	gmst = math.Mod(gmst, 360.0)
	if gmst < 0 {
		gmst += 360.0
	}
	return gmst / 15.0
}

func LST(gmst, lonDeg float64) float64 {
	lst := gmst + lonDeg/15.0
	lst = math.Mod(lst, 24.0)
	if lst < 0 {
		lst += 24.0
	}
	return lst
}

func HourAngle(lstHours, raHours float64) float64 {
	ha := lstHours - raHours
	if ha < 0 {
		ha += 24.0
	}
	return ha * math.Pi / 12.0
}

func AltAz(haRad, decRad, latRad float64) (alt, az float64) {
	sinAlt := math.Sin(decRad)*math.Sin(latRad) +
		math.Cos(decRad)*math.Cos(latRad)*math.Cos(haRad)
	alt = math.Asin(sinAlt)

	cosAz := (math.Sin(decRad) - math.Sin(alt)*math.Sin(latRad)) /
		(math.Cos(alt) * math.Cos(latRad))
	cosAz = math.Max(-1, math.Min(1, cosAz))
	az = math.Acos(cosAz)

	if math.Sin(haRad) > 0 {
		az = 2*math.Pi - az
	}

	alt = alt * 180.0 / math.Pi
	az = az * 180.0 / math.Pi
	return alt, az
}

func GetVisibleStars(catalog StarCatalog, latDeg, lonDeg float64) []VisibleStar {
	now := time.Now().UTC()
	jd := JulianDate(now)
	gmst := GMST(jd)
	lst := LST(gmst, lonDeg)

	latRad := latDeg * math.Pi / 180.0

	var visible []VisibleStar
	for _, s := range catalog.Stars {
		raHours := s.Ra
		decRad := s.Dec * math.Pi / 180.0

		ha := HourAngle(lst, raHours)
		alt, az := AltAz(ha, decRad, latRad)

		if alt > 0 {
			visible = append(visible, VisibleStar{
				Alt:   alt,
				Az:    az,
				Mag:   s.Mag,
				Name:  s.Name,
				Ci:    s.Ci,
				Spect: s.Spect,
			})
		}
	}

	return visible
}
