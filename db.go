package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	_ "github.com/lib/pq"
	"strconv"
	"math/rand"
	"sync"
	"demo/app-1/cache"
	// "demo/app-1/test"
	"demo/app-1/config"
)

var mu sync.Mutex

type Link struct {
	ID          int    `json:"id"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Visits      int    `json:"visits"`
	CreatedAt   string `json:"created_at"`
}

type CreateLinkRequest struct {
    URL string `json:"url"`  // Поле должно называться "url" в JSON
}

func main() {
	//db, err := sql.Open("postgres", "postgres://postgres:1@localhost:5432/golang_db?sslmode=disable")
	config := config.LoadConfig()

	db, err := sql.Open("postgres", config.GetDBConnectionString())

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)

	cache := cache.New()

	http.HandleFunc("GET /links", get3(db))
	http.HandleFunc("POST /links", post1(db))
	http.HandleFunc("GET /links/{short_code}", getOne2(db))
	http.HandleFunc("DELETE /links/{short_code}", delete4(db, cache))
	http.HandleFunc("GET /links/{short_code}/stats", getStat5(db, cache, config))

	// go test.StressTest()

	log.Println("Server started on :"+config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+config.ServerPort, nil))
}

func get3(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(limitStr)
		offsetStr := r.URL.Query().Get("offset")
		offset, err := strconv.Atoi(offsetStr)

		rows, err := db.Query("SELECT id, short_code, original_url, visits, created_at FROM links ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var links []Link
		for rows.Next() {
			var l Link
			rows.Scan(&l.ID, &l.ShortCode, &l.OriginalURL, &l.Visits, &l.CreatedAt)
			links = append(links, l)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)
	}

}

func post1(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			URL string `json:"url"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		mu.Lock()
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		shortCode := make([]byte, 6)
		for i := range shortCode {
			shortCode[i] = charset[rand.Intn(len(charset))]
		}
		mu.Unlock()

		query := `INSERT INTO links (short_code, original_url, visits) 
				  VALUES ($1, $2, 0)`
		
		_, err = db.Exec(query, string(shortCode), req.URL)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"short_code": string(shortCode),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func getOne2(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем short_code из URL path
		shortCode := r.PathValue("short_code")
		
		if shortCode == "" {
			http.Error(w, "short_code is required", http.StatusBadRequest)
			return
		}
		
		// Обновляем счетчик visits и получаем данные за один запрос
		var originalURL string
		var visits int
		
		query := `UPDATE links 
				  SET visits = visits + 1 
				  WHERE short_code = $1 
				  RETURNING original_url, visits`
		
		err := db.QueryRow(query, shortCode).Scan(&originalURL, &visits)
		if err == sql.ErrNoRows {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Формируем ответ
		response := map[string]interface{}{
			"url":    originalURL,
			"visits": visits,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func delete4(db *sql.DB, cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем short_code из URL path
		shortCode := r.PathValue("short_code")
		
		if shortCode == "" {
			http.Error(w, "short_code is required", http.StatusBadRequest)
			return
		}
		
		// Удаляем запись
		query := `DELETE FROM links WHERE short_code = $1`
		
		result, err := db.Exec(query, shortCode)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Проверяем, была ли удалена хоть одна запись
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		if rowsAffected == 0 {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}

		cacheKey := "stats:" + string(shortCode)
        cache.Delete(cacheKey)
		
		// Успешное удаление - возвращаем 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
}

func getStat5(db *sql.DB, cache *cache.Cache, config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем short_code из URL path
		shortCode := r.PathValue("short_code")
		
		if shortCode == "" {
			http.Error(w, "short_code is required", http.StatusBadRequest)
			return
		}

		cacheKey := "stats:" + shortCode
        
        // Пытаемся получить из кеша
        if cachedData, found := cache.Get(cacheKey); found {
            // Данные найдены в кеше - возвращаем их
            w.Header().Set("Content-Type", "application/json")
            // w.Header().Set("X-Cache", "HIT")  // Опционально: пометка что из кеша
            json.NewEncoder(w).Encode(cachedData)
            return
        }
		
		// Получаем статистику без увеличения счетчика visits
		var originalURL string
		var visits int
		var shortCodeFromDB string
		var createdAt string
		
		query := `SELECT short_code, original_url, visits, created_at 
				  FROM links 
				  WHERE short_code = $1`
		
		err := db.QueryRow(query, shortCode).Scan(&shortCodeFromDB, &originalURL, &visits, &createdAt)
		if err == sql.ErrNoRows {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Формируем ответ
		response := map[string]interface{}{
			"short_code": shortCodeFromDB,
			"url":        originalURL,
			"visits":     visits,
			"created_at": createdAt,
		}

		// Сохраняем в кеш на несколько секунд
        cache.Set(cacheKey, response, config.CacheTTL)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}