package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"wishlist-service/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}

type WishlistHandler struct {
	mu             sync.Mutex
	wishlists      map[int64]api.Wishlist
	nextWishlistID int64
	nextItemID     int64
}

func NewWishlistHandler() *WishlistHandler {
	return &WishlistHandler{
		wishlists:      make(map[int64]api.Wishlist),
		nextWishlistID: 1,
		nextItemID:     1,
	}
}

func (h *WishlistHandler) GetWishlists(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]api.Wishlist, 0, len(h.wishlists))
	totalItems := 0
	for _, wl := range h.wishlists {
		result = append(result, wl)
		totalItems += len(wl.Items)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)

	// Бизнес-метрики
	wishlistsGetTotal.Inc()
	itemsReadTotal.Add(float64(totalItems))
	if len(result) > 0 {
		itemsInResponse.Observe(float64(totalItems) / float64(len(result)))
	}

	status := "200"
	if err != nil {
		status = "500"
	}
	IncRequests("getWishlists", "GET", status)
	ObserveRequestDuration("getWishlists", "GET", time.Since(start).Seconds())
}

func (h *WishlistHandler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, span := otel.Tracer("wishlist-service").Start(r.Context(), "wishlist.create")
	defer span.End()

	var req api.CreateWishlistRequest

	_, decodeSpan := otel.Tracer("wishlist-service").Start(ctx, "wishlist.decode_request")
	err := json.NewDecoder(r.Body).Decode(&req)
	decodeSpan.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode error")
		logger.Warn("Failed to decode create wishlist request",
			slog.String("error", err.Error()),
			slog.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		IncRequests("createWishlist", "POST", "400")
		return
	}

	var wishlist api.Wishlist
	var id int64

	_, saveSpan := otel.Tracer("wishlist-service").Start(ctx, "wishlist.save_in_memory")
	h.mu.Lock()

	id = h.nextWishlistID
	h.nextWishlistID++

	wishlist = api.Wishlist{
		Id:    id,
		Name:  req.Name,
		Owner: req.Owner,
		Items: []api.WishlistItem{},
	}

	h.wishlists[id] = wishlist
	updateWishlistsMetric(len(h.wishlists))

	h.mu.Unlock()
	saveSpan.End()

	_, metricsSpan := otel.Tracer("wishlist-service").Start(ctx, "wishlist.update_metrics")
	wishlistsCreatedTotal.Inc()
	metricsSpan.End()

	logger.Info("Wishlist created",
		slog.Int64("wishlist_id", id),
		slog.String("name", req.Name),
		slog.String("owner", req.Owner),
	)

	span.SetAttributes(
		attribute.Int64("wishlist.id", id),
		attribute.String("wishlist.name", req.Name),
		attribute.String("wishlist.owner", req.Owner),
	)

	_, responseSpan := otel.Tracer("wishlist-service").Start(ctx, "wishlist.write_response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(wishlist)
	responseSpan.End()

	status := "201"
	if err != nil {
		status = "500"
		span.RecordError(err)
		span.SetStatus(codes.Error, "response encode error")
	}

	IncRequests("createWishlist", "POST", status)
	ObserveRequestDuration("createWishlist", "POST", time.Since(start).Seconds())
}

func (h *WishlistHandler) GetWishlistById(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	start := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		logger.Warn("Wishlist not found",
			slog.Int64("wishlist_id", wishlistId),
		)
		http.Error(w, "wishlist not found", http.StatusNotFound)
		IncRequests("getWishlistById", "GET", "404")
		return
	}

	// Бизнес-лог
	logger.Info("Wishlist retrieved",
		slog.Int64("wishlist_id", wishlistId),
		slog.String("name", wishlist.Name),
		slog.Int("items_count", len(wishlist.Items)),
	)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(wishlist)

	// Бизнес-метрики
	wishlistsGetTotal.Inc()
	itemsReadTotal.Add(float64(len(wishlist.Items)))
	itemsInResponse.Observe(float64(len(wishlist.Items)))

	status := "200"
	if err != nil {
		status = "500"
	}
	IncRequests("getWishlistById", "GET", status)
	ObserveRequestDuration("getWishlistById", "GET", time.Since(start).Seconds())
}

func (h *WishlistHandler) DeleteWishlist(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	start := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.wishlists[wishlistId]; !ok {
		logger.Warn("Wishlist not found for delete",
			slog.Int64("wishlist_id", wishlistId),
		)
		http.Error(w, "wishlist not found", http.StatusNotFound)
		IncRequests("deleteWishlist", "DELETE", "404")
		return
	}

	logger.Info("Wishlist deleted",
		slog.Int64("wishlist_id", wishlistId),
	)

	delete(h.wishlists, wishlistId)
	updateWishlistsMetric(len(h.wishlists))
	w.WriteHeader(http.StatusNoContent)

	IncRequests("deleteWishlist", "DELETE", "204")
	ObserveRequestDuration("deleteWishlist", "DELETE", time.Since(start).Seconds())
}

func (h *WishlistHandler) AddWishlistItem(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	start := time.Now()
	ctx, span := otel.Tracer("wishlist-service").Start(r.Context(), "wishlist.item.add")
	defer span.End()

	var req api.CreateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode error")
		logger.Warn("Failed to decode add item request",
			slog.String("error", err.Error()),
			slog.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		IncRequests("addWishlistItem", "POST", "400")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		err := fmt.Errorf("wishlist not found: %d", wishlistId)
		span.RecordError(err)
		span.SetStatus(codes.Error, "not found")
		logger.Warn("Wishlist not found for add item",
			slog.Int64("wishlist_id", wishlistId),
		)
		http.Error(w, "wishlist not found", http.StatusNotFound)
		IncRequests("addWishlistItem", "POST", "404")
		return
	}

	itemID := h.nextItemID
	h.nextItemID++

	item := api.WishlistItem{
		Id:      itemID,
		Title:   req.Title,
		Url:     req.Url,
		Price:   req.Price,
		Comment: req.Comment,
	}

	wishlist.Items = append(wishlist.Items, item)
	h.wishlists[wishlistId] = wishlist

	// Бизнес-метрики
	itemsAddedTotal.Inc()

	// Бизнес-лог
	logger.Info("Item added to wishlist",
		slog.Int64("wishlist_id", wishlistId),
		slog.Int64("item_id", itemID),
		slog.String("title", req.Title),
	)

	span.SetAttributes(
		attribute.Int64("wishlist.id", wishlistId),
		attribute.Int64("item.id", itemID),
		attribute.String("item.title", req.Title),
		attribute.Float64("item.price", float64(*req.Price)),
	)

	// Обновляем метрики
	totalItems := 0
	for _, wl := range h.wishlists {
		totalItems += len(wl.Items)
	}
	updateItemsMetric(totalItems)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(item)

	status := "201"
	if err != nil {
		status = "500"
	}
	IncRequests("addWishlistItem", "POST", status)
	ObserveRequestDuration("addWishlistItem", "POST", time.Since(start).Seconds())
	_ = ctx
}

func (h *WishlistHandler) DeleteWishlistItem(w http.ResponseWriter, r *http.Request, wishlistId int64, itemId int64) {
	start := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		logger.Warn("Wishlist not found for delete item",
			slog.Int64("wishlist_id", wishlistId),
		)
		http.Error(w, "wishlist not found", http.StatusNotFound)
		IncRequests("deleteWishlistItem", "DELETE", "404")
		return
	}

	index := -1
	for i, item := range wishlist.Items {
		if item.Id == itemId {
			index = i
			break
		}
	}

	if index == -1 {
		logger.Warn("Item not found for delete",
			slog.Int64("wishlist_id", wishlistId),
			slog.Int64("item_id", itemId),
		)
		http.Error(w, "item not found", http.StatusNotFound)
		IncRequests("deleteWishlistItem", "DELETE", "404")
		return
	}

	logger.Info("Item deleted from wishlist",
		slog.Int64("wishlist_id", wishlistId),
		slog.Int64("item_id", itemId),
	)

	wishlist.Items = append(wishlist.Items[:index], wishlist.Items[index+1:]...)
	h.wishlists[wishlistId] = wishlist

	// Обновляем метрики
	totalItems := 0
	for _, wl := range h.wishlists {
		totalItems += len(wl.Items)
	}
	updateItemsMetric(totalItems)

	w.WriteHeader(http.StatusNoContent)
	IncRequests("deleteWishlistItem", "DELETE", "204")
	ObserveRequestDuration("deleteWishlistItem", "DELETE", time.Since(start).Seconds())
}
