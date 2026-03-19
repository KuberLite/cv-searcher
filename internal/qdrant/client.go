package qdrantclient

import (
	"context"
	"strconv"

	"github.com/qdrant/go-client/qdrant"
)

type Client struct {
	client         *qdrant.Client
	collectionName string
}

type SearchResult struct {
	ProductID int
	Score     float32
	Payload   map[string]*qdrant.Value
}

func New(host string, port int, collectionName string) (*Client, error) {
	qc, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client:         qc,
		collectionName: collectionName,
	}, nil
}

func (c *Client) InitCollection(ctx context.Context, vectorSize int) error {
	collections, err := c.client.ListCollections(ctx)
	if err != nil {
		return err
	}
	for _, collection := range collections {
		if collection == c.collectionName {
			return nil
		}
	}

	return c.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: c.collectionName,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     uint64(vectorSize),
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	})
}

func (c *Client) IndexProduct(ctx context.Context, productID string, vector []float32, payload map[string]interface{}) error {
	id, _ := strconv.ParseUint(productID, 10, 64)
	_, err := c.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: c.collectionName,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(id),
				Vectors: qdrant.NewVectors(vector...),
				Payload: qdrant.NewValueMap(payload),
			},
		},
	})
	return err
}

func (c *Client) Search(ctx context.Context, vector []float32, limit int) ([]SearchResult, error) {
	resp, err := c.client.GetPointsClient().Search(ctx, &qdrant.SearchPoints{
		CollectionName: c.collectionName,
		Vector:         vector,
		Limit:          uint64(limit),
		WithPayload:    &qdrant.WithPayloadSelector{},
	})
	if err != nil {
		return nil, err
	}

	scoredPoints := resp.GetResult()
	results := make([]SearchResult, 0, len(scoredPoints))
	for _, sp := range scoredPoints {
		productID := 0
		if id := sp.GetId(); id != nil {
			productID = int(id.GetNum())
		}
		results = append(results, SearchResult{
			ProductID: productID,
			Score:     sp.GetScore(),
			Payload:   sp.GetPayload(),
		})
	}
	return results, nil
}

func (c *Client) DeleteProduct(ctx context.Context, productID string) error {
	id, _ := strconv.ParseUint(productID, 10, 64)
	_, err := c.client.GetPointsClient().Delete(ctx, &qdrant.DeletePoints{
		CollectionName: c.collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: []*qdrant.PointId{qdrant.NewIDNum(id)},
				},
			},
		},
	})
	return err
}
