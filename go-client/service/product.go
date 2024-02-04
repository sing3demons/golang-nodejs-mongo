package service

import (
	"context"

	"github.com/sing3demons/product/model"
)

type ProductService interface {
	GetProduct(id string) (*model.Products, error)
}

type productService struct {
	productClient ProductServiceClient
}

func NewProductService(productClient ProductServiceClient) ProductService {
	return &productService{productClient}
}

func (s *productService) GetProduct(id string) (*model.Products, error) {
	s.productClient.GetProduct(context.Background(), &ProductRequest{Id: id})
	return nil, nil
}
