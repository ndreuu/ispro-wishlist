package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"wishlist-service/api"
	"wishlist-service/handlers"
)

// responseWriter wrapper для получения статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// setupLogger настраивает slog логгер с выводом в файл и stdout
func setupLogger() *slog.Logger {
	// Создаём директорию для логов
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: could not create log directory: %v", err)
	}
	
	// Создаём файл для логов
	logFile, err := os.OpenFile(logDir+"/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: could not open log file: %v", err)
		// Fallback to stdout only
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Multi-writer: пишем и в файл, и в stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func main() {
	logger := setupLogger()
	
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

	// Middleware для логирования запросов
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         200,
			}
			
			next.ServeHTTP(wrapped, r)
			
			// Логирование запроса
			logger.Info("HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.status),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	})

	// Middleware для сбора метрик
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         200,
			}
			
			next.ServeHTTP(wrapped, r)
			
			// Определяем endpoint
			routePattern := ""
			rctx := chi.RouteContext(r.Context())
			if rctx != nil && rctx.RoutePattern != nil {
				routePattern = rctx.RoutePattern()
			}
			if routePattern == "" {
				routePattern = r.URL.Path
			}
			
			handlers.IncRequests(routePattern, r.Method, strconv.Itoa(wrapped.status))
			handlers.ObserveRequestDuration(routePattern, r.Method, time.Since(start).Seconds())
		})
	})

	h := handlers.NewWishlistHandler()
	handlers.SetLogger(logger)
	api.HandlerFromMux(h, r)

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/wishlist.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	))

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}