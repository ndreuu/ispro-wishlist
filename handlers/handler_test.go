package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"wishlist-service/api"
)

func TestCreateWishlist(t *testing.T) {
	SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

	h := NewWishlistHandler()

	reqBody := api.CreateWishlistRequest{
		Name:  "test-wishlist",
		Owner: "andrew",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/wishlists",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.CreateWishlist(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var wishlist api.Wishlist
	if err := json.NewDecoder(resp.Body).Decode(&wishlist); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if wishlist.Name != reqBody.Name {
		t.Fatalf("expected name %s, got %s", reqBody.Name, wishlist.Name)
	}

	if wishlist.Owner != reqBody.Owner {
		t.Fatalf("expected owner %s, got %s", reqBody.Owner, wishlist.Owner)
	}

	if wishlist.Id == 0 {
		t.Fatal("expected non-zero wishlist id")
	}
}
