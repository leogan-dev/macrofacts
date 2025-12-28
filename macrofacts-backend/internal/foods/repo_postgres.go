package foods

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPostgres struct {
	db *pgxpool.Pool
}

func NewRepoPostgres(db *pgxpool.Pool) *RepoPostgres {
	return &RepoPostgres{db: db}
}

// NOTE: foods_custom schema (from your \d):
// id, created_by_user_id, name, brand, barcode,
// kcal_per_100g, protein_g_per_100g, fat_g_per_100g, carbs_g_per_100g,
// fiber_g_per_100g, sugar_g_per_100g, salt_g_per_100g,
// serving_g, verified, created_at
//
// There is NO serving_size, quantity, sodium_*, saturated_*, mono_*, poly_*, ala_* columns in that output.

func (r *RepoPostgres) Create(ctx context.Context, createdByUserID string, req CreateFoodRequest) (FoodDTO, error) {
	name := strings.TrimSpace(req.Name)

	dto := FoodDTO{
		Source:         FoodSourceCustom,
		Name:           name,
		KcalPer100g:    fptr(req.KcalPer100g),
		ProteinPer100g: fptr(req.ProteinPer100g),
		FatPer100g:     fptr(req.FatPer100g),
		CarbsPer100g:   fptr(req.CarbsPer100g),
		ServingG:       req.ServingG,
		Verified:       false,
	}

	// Optional strings
	if req.Brand != nil {
		b := strings.TrimSpace(*req.Brand)
		if b != "" {
			dto.Brand = &b
		}
	}
	if req.Barcode != nil {
		bc := strings.TrimSpace(*req.Barcode)
		if bc != "" {
			dto.Barcode = &bc
		}
	}

	// Keep these in the DTO only (OFF uses them). We do NOT store them in Postgres.
	if req.ServingSize != nil {
		ss := strings.TrimSpace(*req.ServingSize)
		if ss != "" {
			dto.ServingSize = &ss
		}
	}
	if req.Quantity != nil {
		q := strings.TrimSpace(*req.Quantity)
		if q != "" {
			dto.Quantity = &q
		}
	}

	// Optional nutrients (present in your foods_custom table)
	dto.FiberPer100g = req.FiberPer100g
	dto.SugarPer100g = req.SugarPer100g
	dto.SaltPer100g = req.SaltPer100g

	var id string
	err := r.db.QueryRow(ctx, `
		insert into foods_custom (
			created_by_user_id,
			name, brand, barcode,
			kcal_per_100g, protein_g_per_100g, fat_g_per_100g, carbs_g_per_100g,
			fiber_g_per_100g, sugar_g_per_100g, salt_g_per_100g,
			serving_g
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		returning id::text
	`,
		createdByUserID,
		dto.Name, nullIfEmptyPtr(dto.Brand), nullIfEmptyPtr(dto.Barcode),
		*dto.KcalPer100g, *dto.ProteinPer100g, *dto.FatPer100g, *dto.CarbsPer100g,
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
	if code == "" {
		return nil, nil
	}

	var dto FoodDTO
	dto.Source = FoodSourceCustom

	var brand, barcode *string
	var servingG *float64
	var fiber, sugar, salt *float64

	var kcal, protein, fat, carbs float64
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
		&kcal,
		&protein,
		&fat,
		&carbs,
		&fiber,
		&sugar,
		&salt,
		&servingG,
		&dto.Verified,
	)
	if err != nil {
		// Distinguish not found vs real DB error.
		if isPgNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	dto.KcalPer100g = fptr(kcal)
	dto.ProteinPer100g = fptr(protein)
	dto.FatPer100g = fptr(fat)
	dto.CarbsPer100g = fptr(carbs)

	dto.Brand = brand
	dto.Barcode = barcode
	dto.ServingG = servingG
	dto.FiberPer100g = fiber
	dto.SugarPer100g = sugar
	dto.SaltPer100g = salt

	// Important: ServingSize + Quantity are not stored for custom foods (keep nil).
	dto.ServingSize = nil
	dto.Quantity = nil

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
		   or (brand is not null and lower(brand) like '%' || lower($1) || '%')
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
		dto.Source = FoodSourceCustom

		var brand, barcode *string
		var servingG *float64
		var fiber, sugar, salt *float64

		var kcal, protein, fat, carbs float64
		if err := rows.Scan(
			&dto.ID,
			&dto.Name,
			&brand,
			&barcode,
			&kcal,
			&protein,
			&fat,
			&carbs,
			&fiber,
			&sugar,
			&salt,
			&servingG,
			&dto.Verified,
		); err != nil {
			return nil, err
		}

		dto.KcalPer100g = fptr(kcal)
		dto.ProteinPer100g = fptr(protein)
		dto.FatPer100g = fptr(fat)
		dto.CarbsPer100g = fptr(carbs)

		dto.Brand = brand
		dto.Barcode = barcode
		dto.ServingG = servingG
		dto.FiberPer100g = fiber
		dto.SugarPer100g = sugar
		dto.SaltPer100g = salt

		// Not stored in Postgres for custom foods:
		dto.ServingSize = nil
		dto.Quantity = nil

		out = append(out, dto)
	}

	return out, nil
}

func fptr(v float64) *float64 { return &v }

func isPgNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func nullIfEmptyPtr(s *string) any {
	if s == nil {
		return nil
	}
	if strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}
