package main

import (
	"database/sql"
	"log"
	"net/http"
	_ "github.com/lib/pq"
	"demo/app-1/cache"
	"demo/app-1/config"
	"demo/app-1/handlers"
	"demo/app-1/repository"
	"demo/app-1/usecase"
	// "demo/app-1/test"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)

	cacheClient := cache.New()

	repo := repository.NewLinkRepository(db)
	linkUsecase := usecase.NewLinkUsecase(repo)
	linkHandler := handlers.NewLinkHandler(linkUsecase, cacheClient, cfg)

	http.HandleFunc("GET /links", linkHandler.GetLinks)
	http.HandleFunc("POST /links", linkHandler.PostLink)
	http.HandleFunc("GET /links/{short_code}", linkHandler.GetLink)
	http.HandleFunc("DELETE /links/{short_code}", linkHandler.DeleteLink)
	http.HandleFunc("GET /links/{short_code}/stats", linkHandler.GetStats) 

	// go test.StressTest()

	log.Println("Server started on :" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

