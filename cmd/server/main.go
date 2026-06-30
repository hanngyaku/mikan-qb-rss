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
	"path"
	"path/filepath"
	"strings"

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

	dbPath := env("DB_PATH", "data/app.db")
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	h := handler.New(database, logPath, filepath.Dir(dbPath))
	service.NewRenamer(database).Start(context.Background())
	mux := http.NewServeMux()
	h.Register(mux)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	if webDir := os.Getenv("WEB_DIR"); webDir != "" {
		mux.Handle("/", spaHandler(webDir))
	}

	addr := env("LISTEN_ADDR", ":8081")
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.RequestURI())
		mux.ServeHTTP(w, r)
	})))
}

func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")))
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
