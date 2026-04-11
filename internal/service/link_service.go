package service

import (
	"context"
	"errors"
	"log"

	"LinkStorageService/internal/domain"
	"LinkStorageService/internal/generator"
)

type LinkRepository interface {
	Create(ctx context.Context, link *domain.Link) error
	FindByCode(ctx context.Context, shortCode string) (*domain.Link, error)
	IncrementAndGetVisits(ctx context.Context, shortCode string) (*domain.Link, error)
	IncrementVisitsOnly(ctx context.Context, shortCode string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Link, int64, error)
	Delete(ctx context.Context, shortCode string) error
	Exists(ctx context.Context, shortCode string) (bool, error)
}

type Cache interface {
	Set(ctx context.Context, shortCode string, link *domain.Link) error
	Get(ctx context.Context, shortCode string) (*domain.Link, error)
	Delete(ctx context.Context, shortCode string) error
}

type LinkService struct {
	repo    LinkRepository
	cache   Cache
	gen     *generator.SnowflakeGenerator
	encoder *generator.Base62Encoder
}

func NewLinkService(
	repo LinkRepository,
	cache Cache,
	gen *generator.SnowflakeGenerator,
	encoder *generator.Base62Encoder,
) *LinkService {
	return &LinkService{
		repo:    repo,
		cache:   cache,
		gen:     gen,
		encoder: encoder,
	}
}

func (s *LinkService) Create(ctx context.Context, originalURL string) (string, error) {
	id := s.gen.Generate()
	shortCode := s.encoder.Encode(uint64(id))

	link, err := domain.NewLink(shortCode, originalURL)
	if err != nil {
		return "", err
	}

	if err := s.repo.Create(ctx, link); err != nil {
		return "", err
	}

	if err := s.cache.Set(ctx, shortCode, link); err != nil {
		log.Printf("Failed to cache link: %v", err)
	}

	return shortCode, nil
}

func (s *LinkService) GetByCodeAndIncrement(ctx context.Context, shortCode string) (*domain.Link, error) {
	// 1. Пробуем получить из кеша
	cached, err := s.cache.Get(ctx, shortCode)
	if err != nil {
		log.Printf("Cache get error: %v", err)
	}

	if cached != nil {
		// 2. ФОНОВО увеличиваем счётчик в MongoDB (не блокируем ответ)
		go func() {
			bgCtx := context.Background()
			if err := s.repo.IncrementVisitsOnly(bgCtx, shortCode); err != nil {
				log.Printf("Failed to increment visits for %s: %v", shortCode, err)
			}
		}()

		// 3. Возвращаем данные из кеша (с текущим visits)
		return cached, nil
	}

	// 4. Кеш промах — ОДИН запрос в MongoDB (инкремент + получение сущности)
	link, err := s.repo.IncrementAndGetVisits(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	// 5. Сохраняем в кеш
	go s.cache.Set(context.Background(), shortCode, link)

	return link, nil
}

func (s *LinkService) GetStats(ctx context.Context, shortCode string) (*domain.Link, error) {
	return s.repo.FindByCode(ctx, shortCode)
}

func (s *LinkService) List(ctx context.Context, limit, offset int) ([]*domain.Link, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *LinkService) Delete(ctx context.Context, shortCode string) error {
	exists, err := s.repo.Exists(ctx, shortCode)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("link not found")
	}

	if err := s.repo.Delete(ctx, shortCode); err != nil {
		return err
	}

	go s.cache.Delete(context.Background(), shortCode)

	return nil
}
