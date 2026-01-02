package foods

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPostgresCustom struct {
	db *pgxpool.Pool
}

func NewRepoPostgresCustom(db *pgxpool.Pool) *RepoPostgresCustom {
	return &RepoPostgresCustom{db: db}
}

func (r *RepoPostgresCustom) Create(ctx context.Context, userID string, req CreateFoodRequest) (FoodDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return FoodDTO{}, errors.New("name required")
	}

	// Build OFF-compatible nutriments payload for long-term parity with OFF.
	// We always include the core macros and then merge optional extras.
	nutriments := make(map[string]float64, 16)
	nutriments["energy-kcal_100g"] = req.KcalPer100g
	nutriments["proteins_100g"] = req.ProteinPer100g
	nutriments["carbohydrates_100g"] = req.CarbsPer100g
	nutriments["fat_100g"] = req.FatPer100g

	// Merge explicit optional fields (if provided) into OFF keys.
	if req.FiberPer100g != nil {
		nutriments["fiber_100g"] = *req.FiberPer100g
	}
	if req.SugarPer100g != nil {
		nutriments["sugars_100g"] = *req.SugarPer100g
	}
	if req.SaltPer100g != nil {
		nutriments["salt_100g"] = *req.SaltPer100g
	}
	if req.SodiumPer100g != nil {
		nutriments["sodium_100g"] = *req.SodiumPer100g
	}
	if req.SaturatedFatPer100g != nil {
		nutriments["saturated-fat_100g"] = *req.SaturatedFatPer100g
	}
	if req.MonounsaturatedFatPer100g != nil {
		nutriments["monounsaturated-fat_100g"] = *req.MonounsaturatedFatPer100g
	}
	if req.PolyunsaturatedFatPer100g != nil {
		nutriments["polyunsaturated-fat_100g"] = *req.PolyunsaturatedFatPer100g
	}
	if req.AlphaLinolenicAcidPer100g != nil {
		nutriments["alpha-linolenic-acid_100g"] = *req.AlphaLinolenicAcidPer100g
	}

	// Merge any already OFF-shaped nutriments from the client (preferred extensible format).
	for k, v := range req.Nutriments {
		kk := strings.TrimSpace(k)
		if kk == "" {
			continue
		}
		nutriments[kk] = v
	}

	nutrimentsJSON, err := json.Marshal(nutriments)
	if err != nil {
		return FoodDTO{}, errors.New("failed to encode nutriments")
	}

	var id string
	err = r.db.QueryRow(ctx, `
		insert into foods_custom
			(created_by_user_id, name, brand, barcode,
			 kcal_per_100g, protein_g_per_100g, fat_g_per_100g, carbs_g_per_100g,
			 fiber_g_per_100g, sugar_g_per_100g, salt_g_per_100g,
			 serving_g,
			 nutriments)
		values
			($1, $2, $3, $4,
			 $5, $6, $7, $8,
			 $9, $10, $11,
			 $12,
			 $13)
		returning id::text
	`,
		userID,
		name,
		req.Brand,
		req.Barcode,
		req.KcalPer100g,
		req.ProteinPer100g,
		req.FatPer100g,
		req.CarbsPer100g,
		req.FiberPer100g,
		req.SugarPer100g,
		req.SaltPer100g,
		req.ServingG,
		nutrimentsJSON,
	).Scan(&id)

	if err != nil {
		// Unique constraint for barcode
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				return FoodDTO{}, errors.New("barcode already exists")
			}
		}
		return FoodDTO{}, errors.New("failed to create food")
	}

	return r.ByID(ctx, id)
}

func (r *RepoPostgresCustom) ByID(ctx context.Context, id string) (FoodDTO, error) {
	var dto FoodDTO
	dto.Source = FoodSourceCustom

	var nutrimentsJSON []byte

	err := r.db.QueryRow(ctx, `
		select
			id::text,
			name,
			brand,
			barcode,
			kcal_per_100g::float8,
			protein_g_per_100g::float8,
			carbs_g_per_100g::float8,
			fat_g_per_100g::float8,
			fiber_g_per_100g::float8,
			sugar_g_per_100g::float8,
			salt_g_per_100g::float8,
			serving_g::float8,
			nutriments,
			verified
		from foods_custom
		where id = $1::uuid
	`, id).Scan(
		&dto.ID,
		&dto.Name,
		&dto.Brand,
		&dto.Barcode,
		&dto.KcalPer100g,
		&dto.ProteinPer100g,
		&dto.CarbsPer100g,
		&dto.FatPer100g,
		&dto.FiberPer100g,
		&dto.SugarPer100g,
		&dto.SaltPer100g,
		&dto.ServingG,
		&nutrimentsJSON,
		&dto.Verified,
	)

	if err != nil {
		return FoodDTO{}, errors.New("not found")
	}

	// Map nutriments JSONB into the extended DTO nutrient fields (same keys as OFF).
	if len(nutrimentsJSON) > 0 {
		var nm map[string]any
		if e := json.Unmarshal(nutrimentsJSON, &nm); e == nil {
			dto.SodiumPer100g = pickMaybeFloat(nm, "sodium_100g", "sodium")
			dto.SaturatedFatPer100g = pickMaybeFloat(nm, "saturated-fat_100g", "saturated-fat")
			dto.MonounsaturatedFatPer100g = pickMaybeFloat(nm, "monounsaturated-fat_100g", "monounsaturated-fat")
			dto.PolyunsaturatedFatPer100g = pickMaybeFloat(nm, "polyunsaturated-fat_100g", "polyunsaturated-fat")
			dto.AlphaLinolenicAcidPer100g = pickMaybeFloat(nm, "alpha-linolenic-acid_100g", "alpha-linolenic-acid")
		}
	}

	return dto, nil
}

func (r *RepoPostgresCustom) ByBarcode(ctx context.Context, code string) (*FoodDTO, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("empty barcode")
	}

	var dto FoodDTO
	dto.Source = FoodSourceCustom

	var nutrimentsJSON []byte

	err := r.db.QueryRow(ctx, `
		select
			id::text,
			name,
			brand,
			barcode,
			kcal_per_100g::float8,
			protein_g_per_100g::float8,
			carbs_g_per_100g::float8,
			fat_g_per_100g::float8,
			fiber_g_per_100g::float8,
			sugar_g_per_100g::float8,
			salt_g_per_100g::float8,
			serving_g::float8,
			nutriments,
			verified
		from foods_custom
		where barcode = $1
		limit 1
	`, code).Scan(
		&dto.ID,
		&dto.Name,
		&dto.Brand,
		&dto.Barcode,
		&dto.KcalPer100g,
		&dto.ProteinPer100g,
		&dto.CarbsPer100g,
		&dto.FatPer100g,
		&dto.FiberPer100g,
		&dto.SugarPer100g,
		&dto.SaltPer100g,
		&dto.ServingG,
		&nutrimentsJSON,
		&dto.Verified,
	)

	if err != nil {
		return nil, errors.New("not found")
	}

	if len(nutrimentsJSON) > 0 {
		var nm map[string]any
		if e := json.Unmarshal(nutrimentsJSON, &nm); e == nil {
			dto.SodiumPer100g = pickMaybeFloat(nm, "sodium_100g", "sodium")
			dto.SaturatedFatPer100g = pickMaybeFloat(nm, "saturated-fat_100g", "saturated-fat")
			dto.MonounsaturatedFatPer100g = pickMaybeFloat(nm, "monounsaturated-fat_100g", "monounsaturated-fat")
			dto.PolyunsaturatedFatPer100g = pickMaybeFloat(nm, "polyunsaturated-fat_100g", "polyunsaturated-fat")
			dto.AlphaLinolenicAcidPer100g = pickMaybeFloat(nm, "alpha-linolenic-acid_100g", "alpha-linolenic-acid")
		}
	}

	return &dto, nil
}
