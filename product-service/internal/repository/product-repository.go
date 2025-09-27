package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leandrowiemesfilho/product-service/internal/model"
	"go.uber.org/zap"
)

type ProductRepository interface {
	Create(product *model.CreateProductRequest) (*model.Product, error)
	GetByID(id string) (*model.Product, error)
	GetAll() ([]*model.Product, error)
	Update(id string, product *model.UpdateProductRequest) (*model.Product, error)
	Delete(id string) error
	Exists(id string) (bool, error)
}

type productRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewProductRepository(db *sql.DB, logger *zap.SugaredLogger) ProductRepository {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

func (r *productRepository) Create(req *model.CreateProductRequest) (*model.Product, error) {
	product := &model.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
        INSERT INTO products (id, name, description, price, category, stock, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, name, description, price, category, stock, created_at, updated_at
    `

	err := r.db.QueryRow(
		query,
		product.ID, product.Name, product.Description, product.Price,
		product.Category, product.Stock, product.CreatedAt, product.UpdatedAt,
	).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Category, &product.Stock, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		r.logger.Errorw("Failed to create product", "error", err, "product", product)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	r.logger.Infow("Product created successfully", "product_id", product.ID)
	return product, nil
}

func (r *productRepository) GetByID(id string) (*model.Product, error) {
	query := `SELECT id, name, description, price, category, stock, created_at, updated_at 
              FROM products WHERE id = $1`

	product := &model.Product{}
	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Category, &product.Stock, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warnw("Product not found", "product_id", id)
			return nil, fmt.Errorf("product not found")
		}
		r.logger.Errorw("Failed to get product", "error", err, "product_id", id)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (r *productRepository) GetAll() ([]*model.Product, error) {
	query := `SELECT id, name, description, price, category, stock, created_at, updated_at 
              FROM products ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Errorw("Failed to get products", "error", err)
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		product := &model.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Category, &product.Stock, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorw("Failed to scan product", "error", err)
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorw("Error iterating products", "error", err)
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	r.logger.Infow("Retrieved products successfully", "count", len(products))
	return products, nil
}

func (r *productRepository) Update(id string, req *model.UpdateProductRequest) (*model.Product, error) {
	// First check if product exists
	exists, err := r.Exists(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("product not found")
	}

	query := `
        UPDATE products 
        SET name = COALESCE($1, name),
            description = COALESCE($2, description),
            price = COALESCE($3, price),
            category = COALESCE($4, category),
            stock = COALESCE($5, stock),
            updated_at = $6
        WHERE id = $7
        RETURNING id, name, description, price, category, stock, created_at, updated_at
    `

	product := &model.Product{}
	err = r.db.QueryRow(
		query,
		req.Name, req.Description, req.Price, req.Category, req.Stock,
		time.Now(), id,
	).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Category, &product.Stock, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		r.logger.Errorw("Failed to update product", "error", err, "product_id", id)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	r.logger.Infow("Product updated successfully", "product_id", id)
	return product, nil
}

func (r *productRepository) Delete(id string) error {
	exists, err := r.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product not found")
	}

	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Errorw("Failed to delete product", "error", err, "product_id", id)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	r.logger.Infow("Product deleted successfully", "product_id", id)
	return nil
}

func (r *productRepository) Exists(id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}
	return exists, nil
}
