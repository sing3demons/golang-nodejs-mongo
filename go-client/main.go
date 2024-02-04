package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sing3demons/product/handler"
	"github.com/sing3demons/product/model"
	"github.com/sing3demons/product/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func WriteIds() {
	products, _ := handler.RequestHttpGet[[]model.Products]("http://localhost:8080/products?limit=10000&fields=_id")
	var ids []string
	for _, p := range *products {
		ids = append(ids, p.ID)
	}

	if _, err := os.Stat("id.txt"); err == nil {
		os.Remove("id.txt")
	}

	file, err := os.Create("id.txt")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(strings.Join(ids, ","))
	if err != nil {
		panic(err)
	}
	fmt.Println("Write ids to file", len(ids))
}

func main() {
	// WriteIds()
	file, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	ids, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	productService := service.NewProductServiceClient(cc)
	productHandler := handler.NewProductHandler(productService, string(ids))

	http.HandleFunc("/httpOne", productHandler.GetProductHttpOne)
	http.HandleFunc("/http", productHandler.GetProductHttp)
	http.HandleFunc("/grpcOne", productHandler.GetProductGrpcOne)
	http.HandleFunc("/grpc", productHandler.GetProductGrpc)

	port := "8081"
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Running on port : %s  \n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server is not running : %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server exiting")
}
