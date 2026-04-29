package usecase

import (
	"math/rand"
	"sync"
	"demo/app-1/domain"
	"demo/app-1/repository"
)

var mu sync.Mutex

type LinkUsecase struct {
	repo *repository.LinkRepository
}

func NewLinkUsecase(repo *repository.LinkRepository) *LinkUsecase {
	return &LinkUsecase{repo: repo}
}

func (u *LinkUsecase) generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCode := make([]byte, 6)
	for i := range shortCode {
		shortCode[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortCode)
}

func (u *LinkUsecase) CreateShortLink(req domain.CreateLinkRequest) (string, error) {
	if req.URL == "" {
		return "", nil // error handled in handler
	}

	mu.Lock() // Ваш мьютекс
	shortCode := u.generateShortCode()
	mu.Unlock() // Ваш мьютекс

	err := u.repo.CreateLink(shortCode, req.URL)
	return shortCode, err
}

func (u *LinkUsecase) GetLinkAndRedirect(shortCode string) (string, int, error) {
	return u.repo.GetAndIncrementVisits(shortCode)
}

func (u *LinkUsecase) DeleteLink(shortCode string) (bool, error) {
	rowsAffected, err := u.repo.DeleteLink(shortCode)
	return rowsAffected > 0, err
}

func (u *LinkUsecase) GetLinkStats(shortCode string) (map[string]interface{}, error) {
	shortCodeFromDB, originalURL, visits, createdAt, err := u.repo.GetLinkStats(shortCode)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"short_code": shortCodeFromDB,
		"url":        originalURL,
		"visits":     visits,
		"created_at": createdAt,
	}, nil
}

func (u *LinkUsecase) GetLinks(limit, offset int) ([]domain.Link, error) {
	return u.repo.GetLinks(limit, offset)
}