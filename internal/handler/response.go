package handler

import "github.com/KuberLite/cv-searcher/internal/model"

type SearchResponse struct {
	Hits  []model.Product `json:"hits"`
	Total int             `json:"total"`
	Query string          `json:"query"`
}
