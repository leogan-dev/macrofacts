package foods

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPostgres struct {
	db *pgxpool.Pool
}

func NewRepoPostgres(db *pgxpool.Pool) *RepoPostgres {
	return &RepoPostgres{db: db}
}

func (r *RepoPostgres) Create(ctx context.Context, createdByUserID string, req CreateFoodRequest) (FoodDTO, error) {
	name := strings.TrimSpace(req.Name)

	var dto FoodDTO
	dto.Source = "custom"
	dto.Name = name
	if req.Brand != nil {
		dto.Brand = strings.TrimSpace(*req.Brand)
	}
	if req.Barcode != nil {
		dto.Barcode = strings.TrimSpace(*req.Barcode)
	}

	dto.KcalPer100g = req.KcalPer100g
	dto.ProteinPer100g = req.ProteinPer100g
	dto.FatPer100g = req.FatPer100g
	dto.CarbsPer100g = req.CarbsPer100g
	dto.FiberPer100g = req.FiberPer100g
	dto.SugarPer100g = req.SugarPer100g
	dto.SaltPer100g = req.SaltPer100g
	dto.ServingG = req.ServingG
	dto.Verified = false

	var id string
	err := r.db.QueryRow(ctx, `
		insert into foods_custom (
			created_by_user_id, name, brand, barcode,
			kcal_per_100g, protein_g_per_100g, fat_g_per_100g, carbs_g_per_100g,
			fiber_g_per_100g, sugar_g_per_100g, salt_g_per_100g,
			serving_g
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		returning id::text
	`,
		createdByUserID, dto.Name, nullIfEmpty(dto.Brand), nullIfEmpty(dto.Barcode),
		dto.KcalPer100g, dto.ProteinPer100g, dto.FatPer100g, dto.CarbsPer100g,
		dto.FiberPer100g, dto.SugarPer100g, dto.SaltPer100g,
		dto.ServingG,
	).Scan(&id)
	if err != nil {
		return FoodDTO{}, err
	}

	dto.ID = id
	return dto, nil
}

func (r *RepoPostgres) ByBarcode(ctx context.Context, code string) (*FoodDTO, error) {
	code = strings.TrimSpace(code)

	var dto FoodDTO
	dto.Source = "custom"

	var brand *string
	var barcode *string
	var fiber *float64
	var sugar *float64
	var salt *float64
	var serving *float64

	err := r.db.QueryRow(ctx, `
		select
			id::text,
			name,
			brand,
			barcode,
			kcal_per_100g,
			protein_g_per_100g,
			fat_g_per_100g,
			carbs_g_per_100g,
			fiber_g_per_100g,
			sugar_g_per_100g,
			salt_g_per_100g,
			serving_g,
			verified
		from foods_custom
		where barcode = $1
	`, code).Scan(
		&dto.ID,
		&dto.Name,
		&brand,
		&barcode,
		&dto.KcalPer100g,
		&dto.ProteinPer100g,
		&dto.FatPer100g,
		&dto.CarbsPer100g,
		&fiber,
		&sugar,
		&salt,
		&serving,
		&dto.Verified,
	)
	if err != nil {
		return nil, nil // not found
	}

	if brand != nil {
		dto.Brand = *brand
	}
	if barcode != nil {
		dto.Barcode = *barcode
	}
	dto.FiberPer100g = fiber
	dto.SugarPer100g = sugar
	dto.SaltPer100g = salt
	dto.ServingG = serving

	return &dto, nil
}

func (r *RepoPostgres) Search(ctx context.Context, q string, limit int) ([]FoodDTO, error) {
	q = strings.TrimSpace(q)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	rows, err := r.db.Query(ctx, `
		select
			id::text,
			name,
			brand,
			barcode,
			kcal_per_100g,
			protein_g_per_100g,
			fat_g_per_100g,
			carbs_g_per_100g,
			fiber_g_per_100g,
			sugar_g_per_100g,
			salt_g_per_100g,
			serving_g,
			verified
		from foods_custom
		where lower(name) like '%' || lower($1) || '%'
		order by verified desc, created_at desc
		limit $2
	`, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]FoodDTO, 0, limit)

	for rows.Next() {
		var dto FoodDTO
		dto.Source = "custom"

		var brand *string
		var barcode *string
		var fiber *float64
		var sugar *float64
		var salt *float64
		var serving *float64

		err := rows.Scan(
			&dto.ID,
			&dto.Name,
			&brand,
			&barcode,
			&dto.KcalPer100g,
			&dto.ProteinPer100g,
			&dto.FatPer100g,
			&dto.CarbsPer100g,
			&fiber,
			&sugar,
			&salt,
			&serving,
			&dto.Verified,
		)
		if err != nil {
			return nil, err
		}

		if brand != nil {
			dto.Brand = *brand
		}
		if barcode != nil {
			dto.Barcode = *barcode
		}
		dto.FiberPer100g = fiber
		dto.SugarPer100g = sugar
		dto.SaltPer100g = salt
		dto.ServingG = serving

		out = append(out, dto)
	}

	return out, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
