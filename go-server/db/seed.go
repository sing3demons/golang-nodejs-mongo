package db

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/sing3demons/product/model"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetCategories() []model.Category {
	categories := []model.Category{
		{Name: "Smartphone", ID: "ohxthmFefyB"},
		{Name: "Laptop", ID: "fAtnRvu8Ubm"},
		{Name: "Tablet", ID: "G4Ohz9OAhOu"},
		{Name: "Smartwatch", ID: "RJocTM0yGTl"},
		{Name: "Headphone", ID: "Rbztb13cDAZ"},
		{Name: "Speaker", ID: "xaxUnyN6Qt5"},
		{Name: "Camera", ID: "8US4jyubXCw"},
		{Name: "Printer", ID: "EQlj4ypVGLK"},
		{Name: "Monitor", ID: "8DkTVbS3a"},
		{Name: "Television", ID: "aiQ20tMz9KV"},
	}
	return categories
}

func SeedProducts(productDB *mongo.Collection) {
	start := time.Now()
	fmt.Println("Seeding products...")
	var products []any

	createDb := os.Getenv("CREATE_DB")
	count, err := strconv.Atoi(createDb)
	if err != nil {
		count = 10
	}

	fmt.Printf("Seeding %d products...\n", count)

	categories := GetCategories()

	for i := 0; i < count; i++ {
		name := faker.Name()
		amount := rand.Intn(140) + 10
		lastUpdate := time.Now().UTC()
		category := categories[rand.Intn(len(categories))]
		product := model.Products{
			Type:            "Product",
			Name:            name,
			LifecycleStatus: "active",
			Version:         "1.0",
			LastUpdate:      &lastUpdate,
			ValidFor: &model.ValidFor{
				StartDateTime: time.Now().UTC(),
				EndDateTime:   time.Now().AddDate(1, 0, 0).UTC(),
			}, Category: category,
			ProductPrice: &model.ProductPrice{
				Name:  fmt.Sprintf("Price for %s", name),
				Unit:  "฿",
				Value: float64(amount),
			},
		}
		products = append(products, product)
	}

	insertManyResult, err := productDB.InsertMany(context.Background(), products)
	if err != nil {
		fmt.Println("Error on inserting documents", err)
		return
	}

	fmt.Printf("Seeded %d products id :[%v]\n", count, len(insertManyResult.InsertedIDs))
	fmt.Printf("Seeding took %v\n", time.Since(start).Seconds())
}
