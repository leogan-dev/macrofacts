package foods

type FoodSource string

const (
	FoodSourceOFF    FoodSource = "off"
	FoodSourceCustom FoodSource = "custom"
)

// Canonical food representation used across sources (OFF + custom foods).
type FoodDTO struct {
	ID     string     `json:"id"`
	Source FoodSource `json:"source"`

	// Identity
	Name    string  `json:"name"`
	Brand   *string `json:"brand,omitempty"`
	Barcode *string `json:"barcode,omitempty"`

	// Raw quantities (strings from OFF; custom foods may provide them but we don't persist them unless DB columns exist)
	ServingSize *string `json:"servingSize,omitempty"` // e.g. "30 g"
	Quantity    *string `json:"quantity,omitempty"`    // e.g. "150 gram"

	// Normalized serving size in grams if known (custom foods store this)
	ServingG *float64 `json:"servingG,omitempty"`

	// Core macros per 100g
	// Use pointers so OFF "unknown" stays null (not 0).
	KcalPer100g    *float64 `json:"kcalPer100g"`
	ProteinPer100g *float64 `json:"proteinPer100g"`
	CarbsPer100g   *float64 `json:"carbsPer100g"`
	FatPer100g     *float64 `json:"fatPer100g"`

	// Common secondary nutrients per 100g
	FiberPer100g *float64 `json:"fiberPer100g,omitempty"`
	SugarPer100g *float64 `json:"sugarPer100g,omitempty"`
	SaltPer100g  *float64 `json:"saltPer100g,omitempty"`

	// These are great to expose for OFF, but they are NOT stored in your current foods_custom table.
	// Leave them optional; OFF mapper can fill them, Postgres repo will leave them nil.
	SodiumPer100g *float64 `json:"sodiumPer100g,omitempty"`

	SaturatedFatPer100g       *float64 `json:"saturatedFatPer100g,omitempty"`
	MonounsaturatedFatPer100g *float64 `json:"monounsaturatedFatPer100g,omitempty"`
	PolyunsaturatedFatPer100g *float64 `json:"polyunsaturatedFatPer100g,omitempty"`
	AlphaLinolenicAcidPer100g *float64 `json:"alphaLinolenicAcidPer100g,omitempty"`

	// Custom foods can later be community-verified/moderated
	Verified bool `json:"verified"`
}

// CreateFoodRequest is what the API accepts when users add foods.
// Optional fields are pointers so "unset" is distinct from "set empty".
type CreateFoodRequest struct {
	Name string `json:"name" binding:"required"`

	Brand   *string `json:"brand,omitempty"`
	Barcode *string `json:"barcode,omitempty"`

	ServingSize *string  `json:"servingSize,omitempty"`
	Quantity    *string  `json:"quantity,omitempty"`
	ServingG    *float64 `json:"servingG,omitempty"`

	KcalPer100g    float64 `json:"kcalPer100g" binding:"required"`
	ProteinPer100g float64 `json:"proteinPer100g" binding:"required"`
	CarbsPer100g   float64 `json:"carbsPer100g" binding:"required"`
	FatPer100g     float64 `json:"fatPer100g" binding:"required"`

	FiberPer100g *float64 `json:"fiberPer100g,omitempty"`
	SugarPer100g *float64 `json:"sugarPer100g,omitempty"`
	SaltPer100g  *float64 `json:"saltPer100g,omitempty"`

	// Optional extra nutrients (OFF can provide; custom foods can submit but won't be stored unless you add DB columns)
	SodiumPer100g *float64 `json:"sodiumPer100g,omitempty"`

	SaturatedFatPer100g       *float64 `json:"saturatedFatPer100g,omitempty"`
	MonounsaturatedFatPer100g *float64 `json:"monounsaturatedFatPer100g,omitempty"`
	PolyunsaturatedFatPer100g *float64 `json:"polyunsaturatedFatPer100g,omitempty"`
	AlphaLinolenicAcidPer100g *float64 `json:"alphaLinolenicAcidPer100g,omitempty"`
}
