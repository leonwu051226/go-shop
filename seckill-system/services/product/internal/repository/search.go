package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"seckill-system/pkg/common/config"
	"seckill-system/services/product/internal/model"
)

type ProductSearchRepository struct {
	client *elasticsearch.Client
	index  string
}

type ProductSearchResult struct {
	Products []model.Product
	Total    int64
}

func NewProductSearchRepository(cfg config.ElasticsearchConfig) (*ProductSearchRepository, error) {
	index := strings.TrimSpace(cfg.Index)
	if index == "" {
		index = "products"
	}
	addresses := cfg.Addresses
	if len(addresses) == 0 {
		addresses = []string{"http://localhost:9200"}
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: addresses,
	})
	if err != nil {
		return nil, err
	}

	repo := &ProductSearchRepository{client: client, index: index}
	if err := repo.EnsureIndex(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *ProductSearchRepository) EnsureIndex(ctx context.Context) error {
	res, err := r.client.Indices.Exists([]string{r.index}, r.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}
	if res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("check es index failed: %s", string(body))
	}

	mapping := `{
		"mappings": {
			"properties": {
				"id": { "type": "long" },
				"name": { "type": "text", "analyzer": "standard" },
				"description": { "type": "text", "analyzer": "standard" },
				"price": { "type": "double" },
				"stock": { "type": "integer" },
				"created_at": { "type": "date" },
				"updated_at": { "type": "date" }
			}
		}
	}`
	createRes, err := r.client.Indices.Create(
		r.index,
		r.client.Indices.Create.WithBody(strings.NewReader(mapping)),
		r.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("create es index failed: %s", string(body))
	}
	return nil
}

func (r *ProductSearchRepository) IndexProduct(ctx context.Context, product *model.Product) error {
	doc := model.NewProductDocument(product)
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := r.client.Index(
		r.index,
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(strconv.FormatUint(uint64(product.ID), 10)),
		r.client.Index.WithRefresh("true"),
		r.client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index product failed: %s", string(body))
	}
	return nil
}

func (r *ProductSearchRepository) SearchProducts(ctx context.Context, keyword string, minPrice, maxPrice float64, limit, offset int) (*ProductSearchResult, error) {
	must := make([]map[string]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  keyword,
				"fields": []string{"name^2", "description"},
			},
		})
	} else {
		must = append(must, map[string]any{"match_all": map[string]any{}})
	}

	filters := make([]map[string]any, 0, 1)
	priceRange := map[string]any{}
	if minPrice > 0 {
		priceRange["gte"] = minPrice
	}
	if maxPrice > 0 {
		priceRange["lte"] = maxPrice
	}
	if len(priceRange) > 0 {
		filters = append(filters, map[string]any{"range": map[string]any{"price": priceRange}})
	}

	query := map[string]any{
		"from": offset,
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filters,
			},
		},
		"highlight": map[string]any{
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
			"fields": map[string]any{
				"name":        map[string]any{},
				"description": map[string]any{},
			},
		},
	}
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search products failed: %s", string(body))
	}

	var esResp struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source    model.ProductDocument `json:"_source"`
				Highlight map[string][]string   `json:"highlight"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	products := make([]model.Product, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		doc := hit.Source
		if values := hit.Highlight["name"]; len(values) > 0 {
			doc.Name = values[0]
		}
		if values := hit.Highlight["description"]; len(values) > 0 {
			doc.Description = values[0]
		}
		products = append(products, model.Product{
			ID:          doc.ID,
			Name:        doc.Name,
			Description: doc.Description,
			Price:       doc.Price,
			Stock:       doc.Stock,
			CreatedAt:   doc.CreatedAt,
			UpdatedAt:   doc.UpdatedAt,
		})
	}

	return &ProductSearchResult{Products: products, Total: esResp.Hits.Total.Value}, nil
}
