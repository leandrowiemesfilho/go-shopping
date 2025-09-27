package service

import (
	"github.com/leandrowiemesfilho/product-service/internal/model"
	"github.com/leandrowiemesfilho/product-service/internal/repository"
	"go.uber.org/zap"
)

type ProductService interface {
	CreateProduct(req *model.CreateProductRequest) (*model.Product, error)
	GetProduct(id string) (*model.Product, error)
	GetAllProducts() ([]*model.Product, error)
	UpdateProduct(id string, req *model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(id string) error
}

type productService struct {
	repo   repository.ProductRepository
	logger *zap.SugaredLogger
}

func NewProductService(repo repository.ProductRepository, logger *zap.SugaredLogger) ProductService {
	return &productService{
		repo:   repo,
		logger: logger,
	}
}

func (s *productService) CreateProduct(req *model.CreateProductRequest) (*model.Product, error) {
	if err := req.Validate(); err != nil {
		s.logger.Warnw("Validation failed for create product request", "error", err)
		return nil, err
	}

	product, err := s.repo.Create(req)
	if err != nil {
		s.logger.Errorw("Failed to create product in repository", "error", err)
		return nil, err
	}

	s.logger.Infow("Product created successfully", "product_id", product.ID)
	return product, nil
}

func (s *productService) GetProduct(id string) (*model.Product, error) {
	if id == "" {
		return nil, model.ErrInvalidID
	}

	product, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Errorw("Failed to get product from repository", "error", err, "product_id", id)
		return nil, err
	}

	return product, nil
}

func (s *productService) GetAllProducts() ([]*model.Product, error) {
	products, err := s.repo.GetAll()
	if err != nil {
		s.logger.Errorw("Failed to get all products from repository", "error", err)
		return nil, err
	}

	s.logger.Infow("Retrieved all products", "count", len(products))
	return products, nil
}

func (s *productService) UpdateProduct(id string, req *model.UpdateProductRequest) (*model.Product, error) {
	if id == "" {
		return nil, model.ErrInvalidID
	}

	if err := req.Validate(); err != nil {
		s.logger.Warnw("Validation failed for update product request", "error", err, "product_id", id)
		return nil, err
	}

	product, err := s.repo.Update(id, req)
	if err != nil {
		s.logger.Errorw("Failed to update product in repository", "error", err, "product_id", id)
		return nil, err
	}

	s.logger.Infow("Product updated successfully", "product_id", id)
	return product, nil
}

func (s *productService) DeleteProduct(id string) error {
	if id == "" {
		return model.ErrInvalidID
	}

	if err := s.repo.Delete(id); err != nil {
		s.logger.Errorw("Failed to delete product from repository", "error", err, "product_id", id)
		return err
	}

	s.logger.Infow("Product deleted successfully", "product_id", id)
	return nil
}
