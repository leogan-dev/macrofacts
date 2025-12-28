package foods

import (
	"context"
	"strings"
	"time"
)

type OffRepository interface {
	EnsureTextIndex(ctx context.Context) error
	ByBarcode(ctx context.Context, code string) (*OffFoodDoc, error)
	SearchByNameOrBrand(ctx context.Context, q string, limit int, cursor string) ([]OffFoodDoc, *string, error)
}

type CustomRepository interface {
	Create(ctx context.Context, createdByUserID string, req CreateFoodRequest) (FoodDTO, error)
	ByBarcode(ctx context.Context, code string) (*FoodDTO, error)
	Search(ctx context.Context, q string, limit int) ([]FoodDTO, error)
}

type Service struct {
	offRepo    OffRepository
	customRepo CustomRepository
	offBarcodeCache *barcodeCache
}

func NewService(offRepo OffRepository, customRepo CustomRepository) *Service {
	svc := &Service{offRepo: offRepo, customRepo: customRepo}
	if offRepo != nil {
		svc.offBarcodeCache = newBarcodeCache(10000, 24*time.Hour, 10*time.Minute)
	}
	return svc
}

func (s *Service) EnsureOffIndexes(ctx context.Context) error {
	if s.offRepo == nil {
		return nil
	}
	return s.offRepo.EnsureTextIndex(ctx)
}

// Search: prefer custom foods first, then OFF.
func (s *Service) Search(ctx context.Context, q string, limit int, cursor string) ([]FoodDTO, *string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []FoodDTO{}, nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	out := make([]FoodDTO, 0, limit)

	// 1) custom foods
	if s.customRepo != nil {
		custom, err := s.customRepo.Search(ctx, q, limit)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, custom...)
		if len(out) >= limit {
			return out[:limit], nil, nil
		}
	}

	// 2) OFF foods
	if s.offRepo != nil {
		offDocs, next, err := s.offRepo.SearchByNameOrBrand(ctx, q, limit-len(out), cursor)
		if err != nil {
			return nil, nil, err
		}
		for _, d := range offDocs {
			dto := offDocToDTO(d)
			out = append(out, dto)
			if len(out) >= limit {
				break
			}
		}
		return out, next, nil
	}

	return out, nil, nil
}

// ByBarcode: prefer custom food (user override), else OFF.
func (s *Service) ByBarcode(ctx context.Context, code string) (*FoodDTO, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	// 1) custom
	if s.customRepo != nil {
		dto, err := s.customRepo.ByBarcode(ctx, code)
		if err != nil {
			return nil, err
		}
		if dto != nil {
			return dto, nil
		}
	}

	// Cached OFF lookup
	if s.offBarcodeCache != nil {
		if dto, ok := s.offBarcodeCache.get(code); ok {
			return dto, nil
		}
	}

	// 2) OFF
	if s.offRepo != nil {
		doc, err := s.offRepo.ByBarcode(ctx, code)
		if err != nil {
			if IsNotFound(err) {
				if s.offBarcodeCache != nil {
					s.offBarcodeCache.setNotFound(code)
				}
				return nil, nil
			}
			return nil, err
		}
		dto := offDocToDTO(*doc)
		if s.offBarcodeCache != nil {
			copy := dto
			s.offBarcodeCache.set(code, &copy)
		}
		return &dto, nil
	}

	return nil, nil
}

func (s *Service) CreateCustom(ctx context.Context, userID string, req CreateFoodRequest) (FoodDTO, error) {
	return s.customRepo.Create(ctx, userID, req)
}
