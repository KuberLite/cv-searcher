package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/KuberLite/cv-searcher/internal/model"
	"github.com/meilisearch/meilisearch-go"
)

type Client struct {
	client    meilisearch.ServiceManager
	indexName string
}

func New(host, indexName string) *Client {
	client := meilisearch.New(host)
	return &Client{
		client:    client,
		indexName: indexName,
	}
}

func (c *Client) IndexProduct(ctx context.Context, product model.Product) error {
	documents := []model.Product{product}

	_, err := c.client.Index(c.indexName).AddDocumentsWithContext(ctx, &documents, nil)
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}
	return nil
}

func (c *Client) IndexProducts(ctx context.Context, products []model.Product) error {
	_, err := c.client.Index(c.indexName).AddDocumentsWithContext(ctx, &products, nil)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	return nil
}

func (c *Client) DeleteProduct(ctx context.Context, product model.Product) error {
	identifier := product.ID
	_, err := c.client.Index(c.indexName).DeleteDocumentWithContext(ctx, identifier, nil)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

func (c *Client) Search(ctx context.Context, searchString string, limit int64) ([]model.Product, error) {
	searchRes, err := c.client.Index(c.indexName).SearchWithContext(
		ctx,
		searchString,
		&meilisearch.SearchRequest{
			Limit: limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search [search string: %s]: %w", searchString, err)
	}

	if len(searchRes.Hits) == 0 {
		return []model.Product{}, nil
	}

	hitsBytes, err := json.Marshal(searchRes.Hits)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search hits: %w", err)
	}

	var products []model.Product

	if jsonErr := json.Unmarshal(hitsBytes, &products); jsonErr != nil {
		return nil, fmt.Errorf("JSON parse error: %w", jsonErr)
	}

	return products, nil
}

func (c *Client) ConfigureIndex(ctx context.Context) error {
	searchAttributes := []string{"name", "brand", "description"}

	_, err := c.client.Index(c.indexName).UpdateSearchableAttributesWithContext(ctx, &searchAttributes)

	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
