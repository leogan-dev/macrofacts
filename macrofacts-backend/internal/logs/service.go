package logs

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/auth"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/foods"
)

type Service struct {
	repo    *RepoPostgres
	foods   *foods.Service
	authSvc *auth.Service
}

func NewService(repo *RepoPostgres, foodsSvc *foods.Service, authSvc *auth.Service) *Service {
	return &Service{repo: repo, foods: foodsSvc, authSvc: authSvc}
}

func (s *Service) Today(ctx context.Context, userID string) (TodayResponse, error) {
	if userID == "" {
		return TodayResponse{}, errors.New("unauthorized")
	}

	// Defaults (privacy-first baseline)
	tz := "UTC"
	calGoal := 2000
	pGoal := 150
	cGoal := 200
	fGoal := 70

	// Try to load settings. If it fails for any reason, we still render Today.
	if s.authSvc != nil {
		if settings, err := s.authSvc.GetSettings(userID); err == nil {
			if settings.Timezone != "" {
				tz = settings.Timezone
			}
			if settings.CalorieGoal > 0 {
				calGoal = settings.CalorieGoal
			}
			if settings.ProteinGoalG > 0 {
				pGoal = settings.ProteinGoalG
			}
			if settings.CarbsGoalG > 0 {
				cGoal = settings.CarbsGoalG
			}
			if settings.FatGoalG > 0 {
				fGoal = settings.FatGoalG
			}
		}
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	rows, err := s.repo.ListEntriesForDate(ctx, userID, day)
	if err != nil {
		return TodayResponse{}, errors.New("failed to load entries")
	}

	meals := []TodayMeal{
		{Meal: MealBreakfast},
		{Meal: MealLunch},
		{Meal: MealDinner},
		{Meal: MealSnacks},
	}

	mealIndex := map[Meal]int{
		MealBreakfast: 0,
		MealLunch:     1,
		MealDinner:    2,
		MealSnacks:    3,
	}

	var total MacroTotals

	for _, r := range rows {
		m := Meal(r.Meal)
		i, ok := mealIndex[m]
		if !ok {
			continue
		}

		entry := TodayEntry{
			ID:        r.ID,
			Time:      r.CreatedAt.In(loc).Format("15:04"),
			QuantityG: r.QuantityG,
			Food: TodayEntryFood{
				Name:   r.FoodName,
				Brand:  r.Brand,
				Source: foods.FoodSource(r.Source),
				FoodID: r.FoodID,
				Barcode: func() *string {
					if r.Barcode == nil {
						return nil
					}
					b := *r.Barcode
					return &b
				}(),
			},
			Computed: MacroTotals{
				Calories: r.Calories,
				ProteinG: r.ProteinG,
				CarbsG:   r.CarbsG,
				FatG:     r.FatG,
			},
		}

		meals[i].Entries = append(meals[i].Entries, entry)
		meals[i].Totals.Calories += r.Calories
		meals[i].Totals.ProteinG += r.ProteinG
		meals[i].Totals.CarbsG += r.CarbsG
		meals[i].Totals.FatG += r.FatG

		total.Calories += r.Calories
		total.ProteinG += r.ProteinG
		total.CarbsG += r.CarbsG
		total.FatG += r.FatG
	}

	recentRows, _ := s.repo.ListRecentFoods(ctx, userID, 12)
	recent := make([]RecentFood, 0, len(recentRows))
	for _, r := range recentRows {
		item := RecentFood{
			Source: foods.FoodSource(r.Source),
			FoodID: r.FoodID,
			Barcode: func() *string {
				if r.Barcode == nil {
					return nil
				}
				b := *r.Barcode
				return &b
			}(),
			Name:  r.FoodName,
			Brand: r.Brand,
		}
		item.Per100g.Calories = per100Int(r.Calories, r.QuantityG)
		item.Per100g.ProteinG = per100Float(r.ProteinG, r.QuantityG)
		item.Per100g.CarbsG = per100Float(r.CarbsG, r.QuantityG)
		item.Per100g.FatG = per100Float(r.FatG, r.QuantityG)
		recent = append(recent, item)
	}

	resp := TodayResponse{
		Date:        day.Format("2006-01-02"),
		Meals:       meals,
		RecentFoods: recent,
	}

	resp.Summary.CalorieGoal = calGoal
	resp.Summary.CaloriesConsumed = total.Calories
	resp.Summary.MacrosGoal.ProteinG = pGoal
	resp.Summary.MacrosGoal.CarbsG = cGoal
	resp.Summary.MacrosGoal.FatG = fGoal
	resp.Summary.MacrosConsumed.ProteinG = total.ProteinG
	resp.Summary.MacrosConsumed.CarbsG = total.CarbsG
	resp.Summary.MacrosConsumed.FatG = total.FatG

	return resp, nil
}

func (s *Service) CreateEntry(ctx context.Context, userID string, req CreateEntryRequest) (string, error) {
	if userID == "" {
		return "", errors.New("unauthorized")
	}
	if req.QuantityG <= 0 || req.QuantityG > 5000 {
		return "", errors.New("quantity out of range")
	}
	if req.Meal != MealBreakfast && req.Meal != MealLunch && req.Meal != MealDinner && req.Meal != MealSnacks {
		return "", errors.New("invalid meal")
	}

	// Same default timezone logic as Today()
	tz := "UTC"
	if s.authSvc != nil {
		if settings, err := s.authSvc.GetSettings(userID); err == nil && settings.Timezone != "" {
			tz = settings.Timezone
		}
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	// Resolve food
	var dto *foods.FoodDTO
	switch req.Source {
	case foods.FoodSourceOFF:
		// Accept barcode explicitly, but also allow foodId to act as barcode
		// so name-search selection works without the barcode field.
		var code string
		if req.Barcode != nil && *req.Barcode != "" {
			code = *req.Barcode
		} else if req.FoodID != nil && *req.FoodID != "" {
			code = *req.FoodID
		} else {
			return "", errors.New("barcode required for off")
		}

		// IMPORTANT: persist identifier in `barcode` column.
		// food_id column is UUID and must remain null for OFF entries.
		req.Barcode = &code
		req.FoodID = nil

		dto, err = s.foods.ByBarcode(ctx, code)
		if err != nil || dto == nil {
			return "", errors.New("food not found")
		}

	case foods.FoodSourceCustom:
		if req.FoodID == nil || *req.FoodID == "" {
			return "", errors.New("foodId required for custom")
		}
		d, err2 := s.foods.ByCustomID(ctx, *req.FoodID)
		if err2 != nil {
			return "", errors.New("food not found")
		}
		dto = &d
	default:
		return "", errors.New("invalid source")
	}

	// Compute snapshot (per-100g -> quantity)
	kcal100 := safeNum(dto.KcalPer100g)
	p100 := safeNum(dto.ProteinPer100g)
	c100 := safeNum(dto.CarbsPer100g)
	f100 := safeNum(dto.FatPer100g)

	mult := float64(req.QuantityG) / 100.0
	cal := int(math.Round(kcal100 * mult))
	p := p100 * mult
	ca := c100 * mult
	fa := f100 * mult

	foodName := dto.Name
	brand := dto.Brand
	var foodID *string
	var barcode *string

	if req.Source == foods.FoodSourceCustom {
		foodID = req.FoodID
	} else {
		barcode = req.Barcode
	}

	id, err := s.repo.InsertEntry(
		ctx,
		userID,
		day,
		string(req.Meal),
		string(req.Source),
		foodID,
		barcode,
		foodName,
		brand,
		req.QuantityG,
		cal,
		p,
		ca,
		fa,
	)
	if err != nil {
		return "", errors.New("failed to create entry")
	}
	return id, nil
}

func safeNum(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func per100Int(val int, qty int) int {
	if qty <= 0 {
		return 0
	}
	return int(math.Round(float64(val) * (100.0 / float64(qty))))
}

func per100Float(val float64, qty int) float64 {
	if qty <= 0 {
		return 0
	}
	return val * (100.0 / float64(qty))
}
