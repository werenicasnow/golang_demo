package repository

import (
	"database/sql"
	"demo/app-1/domain"
)

type LinkRepository struct {
	DB *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{DB: db}
}

func (r *LinkRepository) GetLinks(limit, offset int) ([]domain.Link, error) {
	rows, err := r.DB.Query("SELECT id, short_code, original_url, visits, created_at FROM links ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.Link
	for rows.Next() {
		var l domain.Link
		rows.Scan(&l.ID, &l.ShortCode, &l.OriginalURL, &l.Visits, &l.CreatedAt)
		links = append(links, l)
	}
	return links, nil
}

func (r *LinkRepository) CreateLink(shortCode, originalURL string) error {
	query := `INSERT INTO links (short_code, original_url, visits) VALUES ($1, $2, 0)`
	_, err := r.DB.Exec(query, shortCode, originalURL)
	return err
}

func (r *LinkRepository) GetAndIncrementVisits(shortCode string) (string, int, error) {
	var originalURL string
	var visits int
	query := `UPDATE links SET visits = visits + 1 WHERE short_code = $1 RETURNING original_url, visits`
	err := r.DB.QueryRow(query, shortCode).Scan(&originalURL, &visits)
	return originalURL, visits, err
}

func (r *LinkRepository) DeleteLink(shortCode string) (int64, error) {
	result, err := r.DB.Exec(`DELETE FROM links WHERE short_code = $1`, shortCode)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *LinkRepository) GetLinkStats(shortCode string) (string, string, int, string, error) {
	var shortCodeFromDB, originalURL, createdAt string
	var visits int
	query := `SELECT short_code, original_url, visits, created_at FROM links WHERE short_code = $1`
	err := r.DB.QueryRow(query, shortCode).Scan(&shortCodeFromDB, &originalURL, &visits, &createdAt)
	return shortCodeFromDB, originalURL, visits, createdAt, err
}