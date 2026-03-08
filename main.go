package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"wishlist-service/api"
	"wishlist-service/handlers"
)

func main() {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://ndreuu.github.io",
			"http://localhost:8080",
		},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	h := handlers.NewWishlistHandler()
	api.HandlerFromMux(h, r)

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/wishlist.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	))

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}