package foods

import (
	"context"
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

	var id string
	err := r.db.QueryRow(ctx, `
		insert into foods_custom
			(created_by_user_id, name, brand, barcode,
			 kcal_per_100g, protein_g_per_100g, fat_g_per_100g, carbs_g_per_100g,
			 fiber_g_per_100g, sugar_g_per_100g, salt_g_per_100g,
			 serving_g)
		values
			($1, $2, $3, $4,
			 $5, $6, $7, $8,
			 $9, $10, $11,
			 $12)
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
		&dto.Verified,
	)

	if err != nil {
		return FoodDTO{}, errors.New("not found")
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
		&dto.Verified,
	)

	if err != nil {
		return nil, errors.New("not found")
	}
	return &dto, nil
}
