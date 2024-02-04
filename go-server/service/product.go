package service

import (
	context "context"
	"fmt"
	"os"

	"github.com/sing3demons/product/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type productGrpcService struct {
	db *mongo.Collection
}

func NewProductGrpcService(db *mongo.Collection) ProductServiceServer {
	return &productGrpcService{db}
}

func (s *productGrpcService) GetProduct(ctx context.Context, in *ProductRequest) (*ProductResponse, error) {
	id, _ := primitive.ObjectIDFromHex(in.Id)

	var product model.Products
	if err := s.db.FindOne(ctx, bson.M{
		"_id":        id,
		"deleteDate": primitive.Null{},
	}).Decode(&product); err != nil {
		return nil, err
	}

	timestamp := timestamppb.Timestamp{
		Seconds: int64(product.LastUpdate.Second()),
		Nanos:   int32(product.LastUpdate.Nanosecond()),
	}

	return &ProductResponse{
		Id:   product.ID,
		Name: product.Name,
		ProductPrice: &ProductPrice{
			Name:  product.ProductPrice.Name,
			Value: product.ProductPrice.Value,
			Unit:  product.ProductPrice.Unit,
		},
		Type:            product.Type,
		Href:            s.Href(product.ID),
		LifecycleStatus: product.LifecycleStatus,
		Version:         product.Version,
		LastUpdate:      &timestamp,
		ValidFor: &ValidFor{
			StartDateTime: &timestamppb.Timestamp{
				Seconds: int64(product.ValidFor.StartDateTime.Second()),
				Nanos:   int32(product.ValidFor.StartDateTime.Nanosecond()),
			},
			EndDateTime: &timestamppb.Timestamp{
				Seconds: int64(product.ValidFor.EndDateTime.Second()),
				Nanos:   int32(product.ValidFor.EndDateTime.Nanosecond()),
			},
		},
	}, nil
}

func (p *productGrpcService) Href(id string) string {
	return fmt.Sprintf("%s/products/%s", os.Getenv("HOST"), id)
}

func (s *productGrpcService) mustEmbedUnimplementedProductServiceServer() {}
