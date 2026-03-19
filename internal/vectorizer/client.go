package vectorizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	baseUrl    string
	httpClient *http.Client
	cache      map[string][]float32
	cacheMutex sync.RWMutex
}
type EmbeddedRequest struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func New(baseUrl string) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      make(map[string][]float32),
	}
}

func (c *Client) GetVector(ctx context.Context, text string, vectorType string) ([]float32, error) {
	c.cacheMutex.RLock()
	if vec, ok := c.cache[text]; ok {
		c.cacheMutex.RUnlock()
		return vec, nil
	}
	c.cacheMutex.RUnlock()

	reqBody := EmbeddedRequest{Text: text, Type: vectorType}
	reqJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl+"/embed", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vectorizer request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vectorizer request failed with status: %v", resp.StatusCode)
	}

	var result EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	c.cacheMutex.Lock()
	c.cache[text] = result.Embedding
	c.cacheMutex.Unlock()

	return result.Embedding, nil
}

func (c *Client) GetProductVector(ctx context.Context, name, description string) ([]float32, error) {
	text := fmt.Sprintf("%s %s", name, description)
	return c.GetVector(ctx, text, "product")
}
