package handler

import "github.com/KuberLite/cv-searcher/internal/model"

type SearchHit struct {
	model.Product
	Source string `json:"source"`
}

type SearchResponse struct {
	Hits  []SearchHit `json:"hits"`
	Total int         `json:"total"`
	Query string      `json:"query"`
}
