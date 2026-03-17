package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/KuberLite/cv-searcher/internal/meilisearch"
)

type SearchHandler struct {
	searchClient *meilisearch.Client
}

func NewSearchHandler(searchClient *meilisearch.Client) *SearchHandler {
	return &SearchHandler{searchClient: searchClient}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	limit := int64(20)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	products, err := h.searchClient.Search(ctx, q, limit)
	if err != nil {
		http.Error(w, "Search failed [query]: "+q, http.StatusInternalServerError)
		return
	}

	response := SearchResponse{
		Hits:  products,
		Total: len(products),
		Query: q,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
