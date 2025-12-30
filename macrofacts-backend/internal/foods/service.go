package foods

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	offRepo    *RepoMongoOFF
	customRepo *RepoPostgresCustom
	cache      *barcodeCache
}

func NewService(offRepo *RepoMongoOFF, customRepo *RepoPostgresCustom) *Service {
	return &Service{
		offRepo:    offRepo,
		customRepo: customRepo,
		// Big enough to be useful, small enough to be boring
		cache: newBarcodeCache(10000, 24*time.Hour, 30*time.Minute),
	}
}

func (s *Service) Search(ctx context.Context, q string, limit int, cursor string) ([]FoodDTO, *string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []FoodDTO{}, nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	docs, next, err := s.offRepo.SearchByNameOrBrand(ctx, q, limit, cursor)
	if err != nil {
		return nil, nil, err
	}

	out := make([]FoodDTO, 0, len(docs))
	for _, d := range docs {
		dto := offDocToDTO(d)
		out = append(out, dto)
	}
	return out, next, nil
}

func (s *Service) ByBarcode(ctx context.Context, code string) (*FoodDTO, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}

	// Cache hit (including cached not-found)
	if dto, ok := s.cache.get(code); ok {
		return dto, nil
	}

	// 1) OFF first
	doc, err := s.offRepo.ByBarcode(ctx, code)
	if err == nil && doc != nil {
		dto := offDocToDTO(*doc)
		s.cache.set(code, &dto)
		return &dto, nil
	}

	// 2) Custom by barcode
	custom, err2 := s.customRepo.ByBarcode(ctx, code)
	if err2 == nil && custom != nil {
		s.cache.set(code, custom)
		return custom, nil
	}

	s.cache.setNotFound(code)
	return nil, nil
}

func (s *Service) CreateCustom(ctx context.Context, userID string, req CreateFoodRequest) (FoodDTO, error) {
	if userID == "" {
		return FoodDTO{}, errors.New("unauthorized")
	}
	return s.customRepo.Create(ctx, userID, req)
}

func (s *Service) ByCustomID(ctx context.Context, id string) (FoodDTO, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return FoodDTO{}, errors.New("missing id")
	}
	return s.customRepo.ByID(ctx, id)
}
