package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
}

func (r *Repo) CreateUser(ctx context.Context, username string, passwordHash string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", errors.New("username required")
	}

	// Insert defaults so brand-new users always have settings.
	var id string
	err := r.db.QueryRow(ctx, `
		insert into users (
			username,
			password_hash,
			timezone,
			calorie_goal,
			protein_goal_g,
			carbs_goal_g,
			fat_goal_g
		)
		values ($1, $2, 'UTC', 2000, 150, 200, 70)
		returning id::text
	`, username, passwordHash).Scan(&id)

	if err != nil {
		return "", ErrUsernameTaken
	}
	return id, nil
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		select id::text, username, password_hash
		from users
		where username = $1
		limit 1
	`, username).Scan(&u.ID, &u.Username, &u.PasswordHash)

	if err != nil {
		return User{}, errors.New("not found")
	}
	return u, nil
}

func (r *Repo) GetSettings(ctx context.Context, userID string) (MeSettingsResponse, error) {
	var s MeSettingsResponse

	// COALESCE protects us from NULLs for existing rows
	err := r.db.QueryRow(ctx, `
		select
			coalesce(timezone, 'UTC') as timezone,
			coalesce(calorie_goal, 2000) as calorie_goal,
			coalesce(protein_goal_g, 150) as protein_goal_g,
			coalesce(carbs_goal_g, 200) as carbs_goal_g,
			coalesce(fat_goal_g, 70) as fat_goal_g
		from users
		where id = $1::uuid
	`, userID).Scan(
		&s.Timezone,
		&s.CalorieGoal,
		&s.ProteinGoalG,
		&s.CarbsGoalG,
		&s.FatGoalG,
	)

	if err != nil {
		return MeSettingsResponse{}, errors.New("settings not found")
	}
	return s, nil
}

func (r *Repo) UpsertSettings(ctx context.Context, userID string, req UpdateSettingsRequest) error {
	_, err := r.db.Exec(ctx, `
		update users
		set
			timezone       = coalesce($2, timezone, 'UTC'),
			calorie_goal   = coalesce($3, calorie_goal, 2000),
			protein_goal_g = coalesce($4, protein_goal_g, 150),
			carbs_goal_g   = coalesce($5, carbs_goal_g, 200),
			fat_goal_g     = coalesce($6, fat_goal_g, 70)
		where id = $1::uuid
	`, userID, req.Timezone, req.CalorieGoal, req.ProteinGoalG, req.CarbsGoalG, req.FatGoalG)

	return err
}
