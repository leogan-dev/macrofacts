package logs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPostgres struct {
	db *pgxpool.Pool
}

func NewRepoPostgres(db *pgxpool.Pool) *RepoPostgres {
	return &RepoPostgres{db: db}
}

type entryRow struct {
	ID        string
	CreatedAt time.Time
	Meal      string
	Source    string
	FoodID    *string
	Barcode   *string
	FoodName  string
	Brand     *string
	QuantityG int
	Calories  int
	ProteinG  float64
	CarbsG    float64
	FatG      float64
}

func (r *RepoPostgres) InsertEntry(
	ctx context.Context,
	userID string,
	date time.Time,
	meal string,
	source string,
	foodID *string,
	barcode *string,
	foodName string,
	brand *string,
	qtyG int,
	calories int,
	proteinG, carbsG, fatG float64,
) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		insert into food_log_entries
			(user_id, date, meal, source, food_id, barcode, food_name, brand, quantity_g, calories, protein_g, carbs_g, fat_g)
		values
			($1, $2::date, $3, $4, nullif($5,'')::uuid, $6, $7, $8, $9, $10, $11, $12, $13)
		returning id::text
	`,
		userID,
		date.Format("2006-01-02"),
		meal,
		source,
		derefStr(foodID),
		barcode,
		foodName,
		brand,
		qtyG,
		calories,
		proteinG,
		carbsG,
		fatG,
	).Scan(&id)
	return id, err
}

func (r *RepoPostgres) ListEntriesForDate(ctx context.Context, userID string, date time.Time) ([]entryRow, error) {
	rows, err := r.db.Query(ctx, `
		select
			id::text,
			created_at,
			meal,
			source,
			food_id::text,
			barcode,
			food_name,
			brand,
			quantity_g,
			calories,
			protein_g::float8,
			carbs_g::float8,
			fat_g::float8
		from food_log_entries
		where user_id = $1 and date = $2::date
		order by created_at asc
	`, userID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entryRow
	for rows.Next() {
		var e entryRow
		if err := rows.Scan(
			&e.ID,
			&e.CreatedAt,
			&e.Meal,
			&e.Source,
			&e.FoodID,
			&e.Barcode,
			&e.FoodName,
			&e.Brand,
			&e.QuantityG,
			&e.Calories,
			&e.ProteinG,
			&e.CarbsG,
			&e.FatG,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *RepoPostgres) ListRecentFoods(ctx context.Context, userID string, limit int) ([]entryRow, error) {
	rows, err := r.db.Query(ctx, `
		select distinct on (coalesce(barcode, food_id::text))
			id::text,
			created_at,
			meal,
			source,
			food_id::text,
			barcode,
			food_name,
			brand,
			quantity_g,
			calories,
			protein_g::float8,
			carbs_g::float8,
			fat_g::float8
		from food_log_entries
		where user_id = $1
		order by coalesce(barcode, food_id::text), created_at desc
		limit $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entryRow
	for rows.Next() {
		var e entryRow
		if err := rows.Scan(
			&e.ID,
			&e.CreatedAt,
			&e.Meal,
			&e.Source,
			&e.FoodID,
			&e.Barcode,
			&e.FoodName,
			&e.Brand,
			&e.QuantityG,
			&e.Calories,
			&e.ProteinG,
			&e.CarbsG,
			&e.FatG,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
