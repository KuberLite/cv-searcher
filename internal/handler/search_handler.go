package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/KuberLite/cv-searcher/internal/model"
	qdrantclient "github.com/KuberLite/cv-searcher/internal/qdrant"
	"github.com/KuberLite/cv-searcher/internal/vectorizer"
)

const rrfK = 60

type Searcher interface {
	Search(ctx context.Context, query string, limit int64) ([]model.Product, error)
}

type SearchHandler struct {
	searchClient Searcher
	vectorizer   *vectorizer.Client
	qdrantClient *qdrantclient.Client
}

func NewSearchHandler(
	searchClient Searcher,
	vectorizer *vectorizer.Client,
	qdrantClient *qdrantclient.Client,
) *SearchHandler {
	return &SearchHandler{
		searchClient: searchClient,
		vectorizer:   vectorizer,
		qdrantClient: qdrantClient,
	}
}

type rankedHit struct {
	product model.Product
	score   float64
	source  string
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

	// RRF scores and product data keyed by product ID
	rrfScores := make(map[string]float64)
	products := make(map[string]model.Product)
	sources := make(map[string]string)

	// Keyword search (Meilisearch)
	keywordResults, err := h.searchClient.Search(ctx, q, limit)
	if err != nil {
		http.Error(w, "Search failed [query]: "+q, http.StatusInternalServerError)
		return
	}
	for rank, product := range keywordResults {
		rrfScores[product.ID] += 1.0 / float64(rrfK+rank+1)
		products[product.ID] = product
		sources[product.ID] = "keyword"
	}

	// Semantic search (Vectorizer + Qdrant)
	if h.vectorizer != nil && h.qdrantClient != nil {
		queryVec, err := h.vectorizer.GetVector(ctx, q, "query")
		if err != nil {
			log.Printf("[search] vectorizer error: %v", err)
		} else {
			semanticResults, err := h.qdrantClient.Search(ctx, queryVec, int(limit))
			if err != nil {
				log.Printf("[search] Qdrant error: %v", err)
			} else {
				for rank, result := range semanticResults {
					rrfScores[result.ProductID] += 1.0 / float64(rrfK+rank+1)

					if _, exists := products[result.ProductID]; !exists {
						p := model.Product{ID: result.ProductID}
						if v := result.Payload["name"]; v != nil {
							p.Name = v.GetStringValue()
						}
						if v := result.Payload["description"]; v != nil {
							p.Description = v.GetStringValue()
						}
						if v := result.Payload["brand"]; v != nil {
							p.Brand = v.GetStringValue()
						}
						products[result.ProductID] = p
						sources[result.ProductID] = "semantic"
					} else {
						sources[result.ProductID] = "both"
					}
				}
			}
		}
	}

	ranked := make([]rankedHit, 0, len(rrfScores))
	for id, score := range rrfScores {
		ranked = append(ranked, rankedHit{
			product: products[id],
			score:   score,
			source:  sources[id],
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	if int64(len(ranked)) > limit {
		ranked = ranked[:limit]
	}

	hits := make([]SearchHit, len(ranked))
	for i, rh := range ranked {
		hits[i] = SearchHit{
			Product: rh.product,
			Source:  rh.source,
		}
	}

	response := SearchResponse{
		Hits:  hits,
		Total: len(hits),
		Query: q,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
