// @title Mikan qBittorrent RSS API
// @version 0.1
// @description qBittorrent RSS 管理工具 API
// @BasePath /api
package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/example/mikan-qb-rss/docs"
	"github.com/example/mikan-qb-rss/internal/db"
	"github.com/example/mikan-qb-rss/internal/handler"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	database, err := db.Open(env("DB_PATH", "data/app.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	h := handler.New(database)
	mux := http.NewServeMux()
	h.Register(mux)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	addr := env("LISTEN_ADDR", ":8081")
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
