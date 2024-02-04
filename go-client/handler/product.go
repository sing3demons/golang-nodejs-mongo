package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sing3demons/product/model"
	"github.com/sing3demons/product/service"
)

type ProductHandler interface {
	GetProductGrpc(w http.ResponseWriter, r *http.Request)
	GetProductGrpcOne(w http.ResponseWriter, r *http.Request)
	GetProductHttp(w http.ResponseWriter, r *http.Request)
	GetProductHttpOne(w http.ResponseWriter, r *http.Request)
}

type productHandler struct {
	productService service.ProductServiceClient
	ids            string
}

func NewProductHandler(productService service.ProductServiceClient, id string) ProductHandler {
	return &productHandler{productService, id}
}

func (h *productHandler) GetProductHttpOne(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	products := []model.Products{}

	for _, id := range strings.Split(h.ids, ",") {
		p, _ := RequestHttpGet[model.Products]("http://localhost:8080/products/" + id)
		products = append(products, *p)
	}

	durationInMs := time.Since(start).Milliseconds()
	durationFormatted := fmt.Sprintf("%.2f", float64(durationInMs)/1000.0)
	response := map[string]any{
		"durations": durationFormatted + "ms",
		"products":  products,
		"status":    "success",
		"total":     len(products),
	}

	json.NewEncoder(w).Encode(response)
}
func (h *productHandler) GetProductHttp(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	products := []model.Products{}

	var wg sync.WaitGroup
	var mu sync.Mutex
	idList := strings.Split(h.ids, ",")
	fmt.Println("idList", len(idList))
	resultCh := make(chan *model.Products, len(idList))
	poolSize := 10
	semaphore := make(chan struct{}, poolSize)

	for _, id := range idList {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			p, err := RequestHttpGet[model.Products]("http://localhost:8080/products/" + id)
			if err != nil {
				fmt.Println("Error on getting product", err)
				return
			}
			mu.Lock()
			resultCh <- p
			mu.Unlock()

		}(id)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for p := range resultCh {
		products = append(products, *p)
	}

	response := map[string]any{
		"durations": fmt.Sprintf("%.2f ms", float64(time.Since(start).Milliseconds())/1000.0),
		"products":  products[:1000],
		"status":    "success",
		"total":     len(products),
	}

	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) GetProductGrpc(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	idList := strings.Split(h.ids, ",")
	products := []model.Products{}
	var wg sync.WaitGroup

	responseCh := make(chan model.Products, len(idList))
	poolSize := 10
	semaphore := make(chan struct{}, poolSize)
	for _, id := range idList {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			product, err := h.productService.GetProduct(context.Background(), &service.ProductRequest{Id: id})
			if err != nil {
				fmt.Println("Error on getting product", err)
				return
			}

			p := model.Products{
				Type:            product.Type,
				ID:              product.Id,
				Name:            product.Name,
				Href:            product.Href,
				LifecycleStatus: product.LifecycleStatus,
				Version:         product.Version,
				LastUpdate:      product.LastUpdate.AsTime(),
				ValidFor: &model.ValidFor{
					StartDateTime: product.ValidFor.StartDateTime.AsTime(),
					EndDateTime:   product.ValidFor.EndDateTime.AsTime(),
				},
				ProductPrice: &model.ProductPrice{
					Name:  product.ProductPrice.Name,
					Value: product.ProductPrice.Value,
					Unit:  product.ProductPrice.Unit,
				},
			}
			responseCh <- p

		}(id)
	}

	wg.Wait()

	go func() {
		wg.Wait()
		close(responseCh)
	}()
	for data := range responseCh {
		products = append(products, data)
	}

	response := map[string]any{
		"durations": fmt.Sprintf("%.2f ms", float64(time.Since(start).Milliseconds())/1000.0),
		"products":  products[:1000],
		"status":    "success",
		"total":     len(products),
	}

	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) GetProductGrpcOne(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	idList := strings.Split(h.ids, ",")
	products := []model.Products{}

	for _, id := range idList {
		product, err := h.productService.GetProduct(context.Background(), &service.ProductRequest{Id: id})
		if err != nil {
			fmt.Println("Error on getting product", err)
			continue
		}
		products = append(products, model.Products{
			Type:            product.Type,
			ID:              product.Id,
			Name:            product.Name,
			Href:            product.Href,
			LifecycleStatus: product.LifecycleStatus,
			Version:         product.Version,
			LastUpdate:      product.LastUpdate.AsTime(),
			ValidFor: &model.ValidFor{
				StartDateTime: product.ValidFor.StartDateTime.AsTime(),
				EndDateTime:   product.ValidFor.EndDateTime.AsTime(),
			},
			ProductPrice: &model.ProductPrice{
				Name:  product.ProductPrice.Name,
				Value: product.ProductPrice.Value,
				Unit:  product.ProductPrice.Unit,
			},
		})
	}

	response := map[string]any{
		"durations": fmt.Sprintf("%.2f ms", float64(time.Since(start).Milliseconds())/1000.0),
		"products":  products,
		"status":    "success",
		"total":     len(products),
	}

	json.NewEncoder(w).Encode(response)
}

func RequestGrpc(idList []string, client service.ProductServiceClient) (products []model.Products, err error) {
	for _, id := range idList {
		product, err := client.GetProduct(context.Background(), &service.ProductRequest{Id: id})
		if err != nil {
			fmt.Println("Error on getting product", err)
			continue
		}
		products = append(products, model.Products{
			Type:            product.Type,
			ID:              product.Id,
			Name:            product.Name,
			Href:            product.Href,
			LifecycleStatus: product.LifecycleStatus,
			Version:         product.Version,
			LastUpdate:      product.LastUpdate.AsTime(),
			ValidFor: &model.ValidFor{
				StartDateTime: product.ValidFor.StartDateTime.AsTime(),
				EndDateTime:   product.ValidFor.EndDateTime.AsTime(),
			},
			ProductPrice: &model.ProductPrice{
				Name:  product.ProductPrice.Name,
				Value: product.ProductPrice.Value,
				Unit:  product.ProductPrice.Unit,
			},
		})
	}
	return products, nil
}
