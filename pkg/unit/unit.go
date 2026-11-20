package unit

import "strings"

var massFactors = map[string]float64{
	"gram": 1,
	"ons":  100,
	"kg":   1000,
}

var volumeFactors = map[string]float64{
	"ml":    1,
	"liter": 1000,
}

var unitAliases = map[string]string{
	"g":          "gram",
	"gr":         "gram",
	"gram":       "gram",
	"ons":        "ons",
	"kg":         "kg",
	"kilogram":   "kg",
	"ml":         "ml",
	"mililiter":  "ml",
	"milliliter": "ml",
	"l":          "liter",
	"liter":      "liter",
	"bh":         "buah",
	"buah":       "buah",
	"btr":        "butir",
	"butir":      "butir",
	"btg":        "batang",
	"batang":     "batang",
	"bks":        "bungkus",
	"bungkus":    "bungkus",
}

func Canonical(unit string) string {
	u := strings.ToLower(strings.TrimSpace(unit))
	if mapped, ok := unitAliases[u]; ok {
		return mapped
	}
	return u
}

func ConvertQuantity(qty float64, fromUnit, toUnit string) (float64, bool) {
	from := Canonical(fromUnit)
	to := Canonical(toUnit)
	if from == "" || to == "" || from == to {
		return qty, true
	}
	if f1, ok1 := massFactors[from]; ok1 {
		if f2, ok2 := massFactors[to]; ok2 {
			return qty * f1 / f2, true
		}
	}
	if f1, ok1 := volumeFactors[from]; ok1 {
		if f2, ok2 := volumeFactors[to]; ok2 {
			return qty * f1 / f2, true
		}
	}
	return 0, false
}
