package foods

import (
    "strconv"
    "strings"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// OFF -> DTO mapping lives here so service.go stays small and we avoid duplicate helpers.

func offDocToDTO(d OffFoodDoc) FoodDTO {
	dto := FoodDTO{
		Source:   FoodSourceOFF,
		Name:     strings.TrimSpace(d.ProductName),
		Verified: false,
	}

	// Identity
	if code := strings.TrimSpace(d.Code); code != "" {
		dto.Barcode = &code
	}
	if brand := strings.TrimSpace(d.Brands); brand != "" {
		dto.Brand = &brand
	}

	// Raw serving strings
	if s := strings.TrimSpace(d.ServingSize); s != "" {
		dto.ServingSize = &s
	}
	if q := strings.TrimSpace(d.Quantity); q != "" {
		dto.Quantity = &q
	}

	// Core macros per 100g
	dto.KcalPer100g = pickFloatPtr(d.Nutriments,
		"energy-kcal_100g",
		"energy-kcal",
		"energy-kcal_value",
	)
	dto.ProteinPer100g = pickFloatPtr(d.Nutriments, "proteins_100g", "proteins")
	dto.CarbsPer100g = pickFloatPtr(d.Nutriments, "carbohydrates_100g", "carbohydrates")
	dto.FatPer100g = pickFloatPtr(d.Nutriments, "fat_100g", "fat")

	// Secondary but important (per 100g)
	dto.FiberPer100g = pickMaybeFloat(d.Nutriments, "fiber_100g", "fiber")
	dto.SugarPer100g = pickMaybeFloat(d.Nutriments, "sugars_100g", "sugars")
	dto.SaltPer100g = pickMaybeFloat(d.Nutriments, "salt_100g", "salt")
	dto.SodiumPer100g = pickMaybeFloat(d.Nutriments, "sodium_100g", "sodium")

	dto.SaturatedFatPer100g = pickMaybeFloat(d.Nutriments, "saturated-fat_100g", "saturated-fat")
	dto.MonounsaturatedFatPer100g = pickMaybeFloat(d.Nutriments, "monounsaturated-fat_100g", "monounsaturated-fat")
	dto.PolyunsaturatedFatPer100g = pickMaybeFloat(d.Nutriments, "polyunsaturated-fat_100g", "polyunsaturated-fat")
	dto.AlphaLinolenicAcidPer100g = pickMaybeFloat(d.Nutriments, "alpha-linolenic-acid_100g", "alpha-linolenic-acid")

	return dto
}

// -------- helper functions (ONLY define them in ONE file in package foods) --------
// If you already put these in service.go, REMOVE them there and keep them here instead.
// The key is: one copy total.

func pickFloatPtr(m map[string]any, keys ...string) *float64 {
	for _, k := range keys {
		if v, ok := asFloat(m[k]); ok {
			vv := v
			return &vv
		}
	}
	return nil
}

func pickMaybeFloat(m map[string]any, keys ...string) *float64 {
    return pickFloatPtr(m, keys...)
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		ls := strings.ToLower(s)
		if ls == "trace" || ls == "traces" || ls == "<0.1" || ls == "<0,1" {
			return 0, false
		}
		f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case primitive.Decimal128:
		f, err := strconv.ParseFloat(t.String(), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
