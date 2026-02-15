package utils

import (
	"encoding/csv"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Star struct {
	Id    string  `json:"id"`
	Ra    float64 `json:"ra"`
	Dec   float64 `json:"dec"`
	Mag   float64 `json:"mag"`
	Name  string  `json:"name"`
	Con   string  `json:"con"`
	Ci    float64 `json:"ci"`
	Spect string  `json:"spect"`
}

type StarCatalog struct {
	Stars []Star
}

func FilterCatalog(path string) StarCatalog {
	var catalog StarCatalog

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open catalog: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("failed to read catalog: %v", err)
	}

	header := records[0]
	col := make(map[string]int)
	for i, name := range header {
		col[name] = i
	}

	for _, row := range records[1:] {
		mag, err := strconv.ParseFloat(row[col["mag"]], 64)
		if err != nil {
			continue
		}
		if mag <= 15.0 {
			continue
		}

		ra, _ := strconv.ParseFloat(row[col["ra"]], 64)
		dec, _ := strconv.ParseFloat(row[col["dec"]], 64)
		ci, _ := strconv.ParseFloat(row[col["ci"]], 64)

		catalog.Stars = append(catalog.Stars, Star{
			Id:    row[col["id"]],
			Ra:    ra,
			Dec:   dec,
			Mag:   mag,
			Name:  row[col["proper"]],
			Con:   row[col["con"]],
			Ci:    ci,
			Spect: row[col["spect"]],
		})
	}

	return catalog
}
