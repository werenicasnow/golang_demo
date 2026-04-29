package domain

type Link struct {
	ID          int    `json:"id"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Visits      int    `json:"visits"`
	CreatedAt   string `json:"created_at"`
}

type CreateLinkRequest struct {
    URL string `json:"url"`
}