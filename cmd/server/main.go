// @title Mikan qBittorrent RSS API
// @version 0.1
// @description qBittorrent RSS 管理工具 API
// @BasePath /api
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/example/mikan-qb-rss/docs"
	"github.com/example/mikan-qb-rss/internal/db"
	"github.com/example/mikan-qb-rss/internal/handler"
	"github.com/example/mikan-qb-rss/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	logPath := env("LOG_PATH", "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	database, err := db.Open(env("DB_PATH", "data/app.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	h := handler.New(database, logPath)
	service.NewRenamer(database).Start(context.Background())
	mux := http.NewServeMux()
	h.Register(mux)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	addr := env("LISTEN_ADDR", ":8081")
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.RequestURI())
		mux.ServeHTTP(w, r)
	})))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
