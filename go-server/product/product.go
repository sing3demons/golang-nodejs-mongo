package product

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sing3demons/product/db"
	"github.com/sing3demons/product/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

type IProduct interface {
	GetCategories(c *gin.Context)
	FindProducts(c *gin.Context)
	FindProductByID(c *gin.Context)
	FindProductByCategoryID(c *gin.Context)
}

type ProductHandler struct {
	collection *mongo.Collection
}

func NewProductHandler(collection *mongo.Collection) IProduct {
	return &ProductHandler{collection}
}

func (p *ProductHandler) GetCategories(c *gin.Context) {
	categories := []model.Category{}
	for _, category := range db.GetCategories() {
		category.Href = p.HrefCategory(category.ID)
		category.ProductTotal = p.CountProduct(category.ID)
		// category.Products = p.ProductId(category.ID)
		categories = append(categories, category)
	}

	fmt.Println(len(p.Expands(c.Query("expand"))))

	c.JSON(http.StatusOK, categories)
}

func (p *ProductHandler) ProductId(id string) []model.Products {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var products []model.Products
	filter := bson.D{{Key: "category._id", Value: id}, {Key: "deleteDate", Value: primitive.Null{}}}
	opts := options.FindOptions{}
	opts.SetProjection(bson.M{"_id": 1, "name": 1})
	cur, _ := p.collection.Find(ctx, filter, &opts)

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var product model.Products
		cur.Decode(&product)
		products = append(products, model.Products{
			ID:   product.ID,
			Name: product.Name,
			Href: p.Href(product.ID),
		})
	}
	return products

}

func (p *ProductHandler) Expands(expand string) []string {
	if expand != "" {
		return strings.Split(expand, ",")
	}
	return nil
}
func (p *ProductHandler) CountProduct(categoryId string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, _ := p.collection.CountDocuments(ctx, bson.D{
		{Key: "category._id", Value: categoryId}, {Key: "deleteDate", Value: primitive.Null{}},
	})
	return count
}

func (p *ProductHandler) FindProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	products := []model.Products{}

	filter := filterOption(c)
	opts := FindOptions(c)
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusOK, products)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var product model.Products
		if err := cursor.Decode(&product); err != nil {
			c.JSON(500, gin.H{
				"message": "Error on decoding documents",
			})
			return
		}

		if product.Category.ID != "" {
			product.Category.Href = p.HrefCategory(product.Category.ID)
		}

		product.Href = p.Href(product.ID)
		if product.ValidFor != nil {
			loc, _ := time.LoadLocation("Asia/Bangkok")
			if !product.ValidFor.StartDateTime.IsZero() {
				product.ValidFor.StartDateTime = product.ValidFor.StartDateTime.In(loc)
			}
			if !product.ValidFor.EndDateTime.IsZero() {
				product.ValidFor.EndDateTime = product.ValidFor.EndDateTime.In(loc)
			}
		}

		products = append(products, product)
	}

	c.JSON(200, products)
}

func filterOption(c *gin.Context) primitive.D {
	filter := bson.D{bson.E{Key: "deleteDate", Value: primitive.Null{}}}
	id := c.Query("id")
	if id != "" {
		var ids []primitive.ObjectID
		for _, hex := range strings.Split(id, ",") {
			objectID, err := primitive.ObjectIDFromHex(hex)
			if err != nil {
				continue
			}
			ids = append(ids, objectID)
		}
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}})
	}

	if name := c.Query("name"); name != "" {
		filter = append(filter, bson.E{Key: "name", Value: bson.D{{Key: "$regex", Value: name}}})
	}

	if category := c.Param("categoryId"); category != "" {
		filter = append(filter, bson.E{Key: "category._id", Value: category})
	}
	return filter
}

func FindOptions(c *gin.Context) *options.FindOptions {
	sLimit := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(sLimit)
	field := c.Query("fields")
	var fields []string
	projection := bson.M{}
	if field != "" {
		fields = strings.Split(field, ",")
	}
	opts := options.FindOptions{}
	opts.SetLimit(int64(limit))
	if len(fields) != 0 {
		for _, f := range fields {
			projection[f] = 1
		}
		opts.SetProjection(projection)
	}
	return &opts
}

func (p *ProductHandler) FindProductByID(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	product := model.Products{}
	productId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusFound, product)
		return
	}

	filter := bson.M{"_id": productId, "deleteDate": nil}

	if err := p.collection.FindOne(ctx, filter).Decode(&product); err != nil {
		c.JSON(http.StatusFound, product)
		return
	}
	if product.Category.ID != "" {
		product.Category.Href = p.HrefCategory(product.Category.ID)
	}

	product.Href = p.Href(product.ID)

	c.JSON(200, product)
}

func (p *ProductHandler) FindProductByCategoryID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	products := []model.Products{}

	filter := filterOption(c)
	opts := FindOptions(c)
	cur, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusOK, products)
		return
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var product model.Products
		if err := cur.Decode(&product); err != nil {
			c.JSON(500, gin.H{
				"message": "Error on decoding documents",
			})
			return
		}
		if product.Category.ID != "" {
			product.Category.Href = fmt.Sprintf("%s/categories/%s", os.Getenv("HOST"), product.Category.ID)
		}
		product.Href = p.Href(product.ID)
		products = append(products, product)
	}

	c.JSON(200, products)
}

func (p *ProductHandler) Href(id string) string {
	return fmt.Sprintf("%s/products/%s", os.Getenv("HOST"), id)
}

func (p *ProductHandler) HrefCategory(id string) string {
	return fmt.Sprintf("%s/categories/%s", os.Getenv("HOST"), id)
}
