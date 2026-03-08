package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"wishlist-service/api"
)

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
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]api.Wishlist, 0, len(h.wishlists))
	for _, wl := range h.wishlists {
		result = append(result, wl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *WishlistHandler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	var req api.CreateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.nextWishlistID
	h.nextWishlistID++

	wishlist := api.Wishlist{
		Id:    id,
		Name:  req.Name,
		Owner: req.Owner,
		Items: []api.WishlistItem{},
	}

	h.wishlists[id] = wishlist

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(wishlist)
}

func (h *WishlistHandler) GetWishlistById(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		http.Error(w, "wishlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(wishlist)
}

func (h *WishlistHandler) DeleteWishlist(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.wishlists[wishlistId]; !ok {
		http.Error(w, "wishlist not found", http.StatusNotFound)
		return
	}

	delete(h.wishlists, wishlistId)
	w.WriteHeader(http.StatusNoContent)
}

func (h *WishlistHandler) AddWishlistItem(w http.ResponseWriter, r *http.Request, wishlistId int64) {
	var req api.CreateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		http.Error(w, "wishlist not found", http.StatusNotFound)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *WishlistHandler) DeleteWishlistItem(w http.ResponseWriter, r *http.Request, wishlistId int64, itemId int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	wishlist, ok := h.wishlists[wishlistId]
	if !ok {
		http.Error(w, "wishlist not found", http.StatusNotFound)
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
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	wishlist.Items = append(wishlist.Items[:index], wishlist.Items[index+1:]...)
	h.wishlists[wishlistId] = wishlist

	w.WriteHeader(http.StatusNoContent)
}