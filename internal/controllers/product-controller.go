package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/grpc-products-service/pb"
	"github.com/yaninyzwitty/grpc-products-service/snowflake"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductController struct {
	session *gocql.Session
	pb.UnimplementedProductsServiceServer
}

func NewProductController(session *gocql.Session) *ProductController {
	return &ProductController{session: session}
}

func (c *ProductController) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	if req.Name == "" || req.Description == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name and description are required")
	}
	createCategoryQuery := `INSERT INTO chat.categories(id, name, description, created_at) VALUES(?, ?, ?, ?)`

	categoryId, err := snowflake.GenerateID()
	if err != nil {
		return nil,
			status.Errorf(codes.Internal, "Failed to generate category id: %v", err)
	}

	now := time.Now()
	if err := c.session.Query(
		createCategoryQuery,
		categoryId,
		req.Name,
		req.Description,
		now,
	).WithContext(ctx).Exec(); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create category")
	}

	return &pb.CreateCategoryResponse{
		Category: &pb.Category{
			Id:          int64(categoryId),
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   timestamppb.New(now),
		},
	}, nil
}
func (c *ProductController) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	if req.CategoryId <= 0 || req.Name == "" || req.Description == "" || req.Price <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid product data")

	}

	productID, err := snowflake.GenerateID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate product ID: %v", err)
	}

	outboxID := gocql.TimeUUID()
	bucket := time.Now().Format("2006-01-02")
	now := time.Now()
	eventType := "CREATE_PRODUCT"

	product := &pb.Product{
		Id:          int64(productID),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryId:  req.CategoryId,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}

	payload, err := json.Marshal(product)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to marshal product data: %v", err)
	}

	batch := c.session.NewBatch(gocql.LoggedBatch)
	batch.WithContext(ctx)

	batch.Query(
		`INSERT INTO chat.products 
		(id, name, description, price, stock, category_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		product.Id, product.Name, product.Description, product.Price,
		product.Stock, product.CategoryId, now, now,
	)

	// Add outbox event query
	batch.Query(
		`INSERT INTO chat.products_outbox 
		(id, bucket, payload, event_type) 
		VALUES (?, ?, ?, ?)`,
		outboxID, bucket, payload, eventType,
	)

	// execute batch
	if err := c.session.ExecuteBatch(batch); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create product: %v", err)
	}

	return &pb.CreateProductResponse{
		Product: product,
	}, nil

}
