package logs

import "github.com/leogan-dev/macrofacts/macrofacts-backend/internal/foods"

type Meal string

const (
	MealBreakfast Meal = "breakfast"
	MealLunch     Meal = "lunch"
	MealDinner    Meal = "dinner"
	MealSnacks    Meal = "snacks"
)

type MacroTotals struct {
	Calories int     `json:"calories"`
	ProteinG float64 `json:"protein_g"`
	CarbsG   float64 `json:"carbs_g"`
	FatG     float64 `json:"fat_g"`
}

type TodaySummary struct {
	CalorieGoal      int `json:"calorieGoal"`
	CaloriesConsumed int `json:"caloriesConsumed"`
	MacrosGoal       struct {
		ProteinG int `json:"protein_g"`
		CarbsG   int `json:"carbs_g"`
		FatG     int `json:"fat_g"`
	} `json:"macrosGoal"`
	MacrosConsumed struct {
		ProteinG float64 `json:"protein_g"`
		CarbsG   float64 `json:"carbs_g"`
		FatG     float64 `json:"fat_g"`
	} `json:"macrosConsumed"`
}

type TodayEntryFood struct {
	Name    string           `json:"name"`
	Brand   *string          `json:"brand,omitempty"`
	Source  foods.FoodSource `json:"source"`
	FoodID  *string          `json:"foodId,omitempty"`
	Barcode *string          `json:"barcode,omitempty"`
}

type TodayEntry struct {
	ID        string         `json:"id"`
	Time      string         `json:"time"`
	Food      TodayEntryFood `json:"food"`
	QuantityG int            `json:"quantity_g"`
	Computed  MacroTotals    `json:"computed"`
}

type TodayMeal struct {
	Meal    Meal         `json:"meal"`
	Totals  MacroTotals  `json:"totals"`
	Entries []TodayEntry `json:"entries"`
}

type RecentFood struct {
	Source  foods.FoodSource `json:"source"`
	FoodID  *string          `json:"foodId,omitempty"`
	Barcode *string          `json:"barcode,omitempty"`
	Name    string           `json:"name"`
	Brand   *string          `json:"brand,omitempty"`
	Per100g struct {
		Calories int     `json:"calories"`
		ProteinG float64 `json:"protein_g"`
		CarbsG   float64 `json:"carbs_g"`
		FatG     float64 `json:"fat_g"`
	} `json:"per100g"`
	Serving *struct {
		Label string  `json:"label"`
		Grams float64 `json:"grams"`
	} `json:"serving,omitempty"`
}

type TodayResponse struct {
	Date        string       `json:"date"`
	Summary     TodaySummary `json:"summary"`
	Meals       []TodayMeal  `json:"meals"`
	RecentFoods []RecentFood `json:"recentFoods"`
}

type CreateEntryRequest struct {
	Meal      Meal             `json:"meal" binding:"required"`
	Source    foods.FoodSource `json:"source" binding:"required"`
	FoodID    *string          `json:"foodId,omitempty"`
	Barcode   *string          `json:"barcode,omitempty"`
	QuantityG int              `json:"quantity_g" binding:"required"`
}

type CreateEntryResponse struct {
	ID string `json:"id"`
}
