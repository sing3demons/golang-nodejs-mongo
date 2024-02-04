package main

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sing3demons/product/db"
	"github.com/sing3demons/product/product"
	"github.com/sing3demons/product/service"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(".env.dev"); err != nil {
		panic(err)
	}
	connect := db.New("product")
	defer connect.Disconnect()
	productCol := connect.Collection("products")
	// db.SeedProducts(productCol)

	productHandler := product.NewProductHandler(productCol)

	s := grpc.NewServer()

	go func() {
		lis, err := net.Listen("tcp", ":"+"50051")
		if err != nil {
			panic(err)
		}

		service.RegisterProductServiceServer(s, service.NewProductGrpcService(productCol))

		fmt.Printf("gRPC server is running at %s\n", lis.Addr().String())
		s.Serve(lis)
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/products", productHandler.FindProducts)
	r.GET("/products/:id", productHandler.FindProductByID)

	r.Run(":8080")
}
