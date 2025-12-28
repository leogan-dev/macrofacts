package foods

type FoodDTO struct {
	Source  string `json:"source"` // "custom" (later "off")
	ID      string `json:"id,omitempty"`
	Barcode string `json:"barcode,omitempty"`

	Name  string `json:"name"`
	Brand string `json:"brand,omitempty"`

	KcalPer100g    float64 `json:"kcalPer100g"`
	ProteinPer100g float64 `json:"proteinPer100g"`
	FatPer100g     float64 `json:"fatPer100g"`
	CarbsPer100g   float64 `json:"carbsPer100g"`

	FiberPer100g *float64 `json:"fiberPer100g,omitempty"`
	SugarPer100g *float64 `json:"sugarPer100g,omitempty"`
	SaltPer100g  *float64 `json:"saltPer100g,omitempty"`

	ServingG *float64 `json:"servingG,omitempty"`
	Verified bool     `json:"verified"`
}

type CreateFoodRequest struct {
	Name    string  `json:"name"`
	Brand   *string `json:"brand"`
	Barcode *string `json:"barcode"`

	KcalPer100g    float64 `json:"kcalPer100g"`
	ProteinPer100g float64 `json:"proteinPer100g"`
	FatPer100g     float64 `json:"fatPer100g"`
	CarbsPer100g   float64 `json:"carbsPer100g"`

	FiberPer100g *float64 `json:"fiberPer100g"`
	SugarPer100g *float64 `json:"sugarPer100g"`
	SaltPer100g  *float64 `json:"saltPer100g"`

	ServingG *float64 `json:"servingG"`
}
