package controllers

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/grpc-products-service/pb"
)

type ProductController struct {
	session *gocql.Session
	pb.UnimplementedProductsServiceServer
}

func NewProductController(session *gocql.Session) *ProductController {
	return &ProductController{session: session}
}
func (c *ProductController) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	return &pb.CreateProductResponse{}, nil
}
