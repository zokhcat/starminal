package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"
)

func SaveScreenshot(stars []VisibleStar, width, height int, pincode string) (string, error) {
	const scale = 4
	imgW := width * scale
	imgH := height * scale
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	bg := color.RGBA{10, 10, 30, 255}
	for y := range imgH {
		for x := range imgW {
			img.Set(x, y, bg)
		}
	}

	centerX := float64(imgW) / 2.0
	centerY := float64(imgH) / 2.0
	radius := math.Min(centerX, centerY)

	for _, s := range stars {
		altRad := s.Alt * math.Pi / 180.0
		azRad := s.Az * math.Pi / 180.0

		r := math.Cos(altRad) / (1.0 + math.Sin(altRad))
		px := r * math.Sin(azRad)
		py := -r * math.Cos(azRad)

		cx := int(centerX + px*radius)
		cy := int(centerY + py*radius)

		if cx < 0 || cx >= imgW || cy < 0 || cy >= imgH {
			continue
		}

		col := starColor(s.Ci)
		dotR := starDotRadius(s.Mag)

		for dy := -dotR; dy <= dotR; dy++ {
			for dx := -dotR; dx <= dotR; dx++ {
				if dx*dx+dy*dy <= dotR*dotR {
					sx, sy := cx+dx, cy+dy
					if sx >= 0 && sx < imgW && sy >= 0 && sy < imgH {
						img.Set(sx, sy, col)
					}
				}
			}
		}
	}

	filename := fmt.Sprintf("starminal_%s_%s.png", pincode, time.Now().Format("20060102_150405"))
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", err
	}

	return filename, nil
}

func starColor(ci float64) color.RGBA {
	switch {
	case math.IsNaN(ci) || ci == 0:
		return color.RGBA{255, 255, 255, 255}
	case ci < -0.1:
		return color.RGBA{130, 180, 255, 255}
	case ci < 0.3:
		return color.RGBA{180, 210, 255, 255}
	case ci < 0.6:
		return color.RGBA{255, 255, 255, 255}
	case ci < 1.0:
		return color.RGBA{255, 240, 180, 255}
	case ci < 1.5:
		return color.RGBA{255, 190, 100, 255}
	default:
		return color.RGBA{255, 100, 80, 255}
	}
}

func starDotRadius(mag float64) int {
	switch {
	case mag < 3.0:
		return 4
	case mag < 5.0:
		return 2
	default:
		return 1
	}
}
