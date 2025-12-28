package foods

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *RepoPostgres
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{repo: NewRepoPostgres(db)}
}

func (s *Service) Create(ctx context.Context, createdByUserID string, req CreateFoodRequest) (FoodDTO, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return FoodDTO{}, errors.New("name is required")
	}

	if req.KcalPer100g < 0 || req.ProteinPer100g < 0 || req.FatPer100g < 0 || req.CarbsPer100g < 0 {
		return FoodDTO{}, errors.New("macros must be non-negative")
	}

	// Optional barcode normalization
	if req.Barcode != nil {
		b := strings.TrimSpace(*req.Barcode)
		if b == "" {
			req.Barcode = nil
		} else {
			req.Barcode = &b
		}
	}

	return s.repo.Create(ctx, createdByUserID, req)
}

func (s *Service) Search(ctx context.Context, q string, limit int) ([]FoodDTO, error) {
	if strings.TrimSpace(q) == "" {
		return []FoodDTO{}, nil
	}
	return s.repo.Search(ctx, q, limit)
}

func (s *Service) ByBarcode(ctx context.Context, code string) (*FoodDTO, error) {
	if strings.TrimSpace(code) == "" {
		return nil, nil
	}
	return s.repo.ByBarcode(ctx, code)
}
