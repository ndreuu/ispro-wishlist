package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Метрики вишлистов
	wishlistsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wishlist_service_wishlists_total",
		Help: "Total number of wishlists",
	})

	// Метрики элементов вишлистов
	wishlistItemsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wishlist_service_items_total",
		Help: "Total number of wishlist items",
	})

	// Бизнес-метрики операций
	wishlistsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wishlist_service_wishlists_created_total",
		Help: "Total number of wishlists created",
	})

	wishlistsGetTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wishlist_service_wishlists_get_total",
		Help: "Total number of wishlist read operations",
	})

	itemsAddedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wishlist_service_items_added_total",
		Help: "Total number of items added to wishlists",
	})

	itemsReadTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wishlist_service_items_read_total",
		Help: "Total number of items returned in responses",
	})

	// Распределение количества элементов в ответе
	itemsInResponse = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "wishlist_service_items_in_response",
		Help:    "Distribution of items count in API responses",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
	})

	// Метрики запросов
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wishlist_service_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"endpoint", "method", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "wishlist_service_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"endpoint", "method"})
)

// updateWishlistsMetric обновляет метрику количества вишлистов
func updateWishlistsMetric(count int) {
	wishlistsTotal.Set(float64(count))
}

// updateItemsMetric обновляет метрику количества элементов
func updateItemsMetric(count int) {
	wishlistItemsTotal.Set(float64(count))
}

// ObserveRequestDuration наблюдает за длительностью запроса
func ObserveRequestDuration(endpoint, method string, duration float64) {
	httpRequestDuration.WithLabelValues(endpoint, method).Observe(duration)
}

// IncRequests увеличивает счетчик запросов
func IncRequests(endpoint, method, status string) {
	httpRequestsTotal.WithLabelValues(endpoint, method, status).Inc()
}
