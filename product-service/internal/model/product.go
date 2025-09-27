package model

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Description string    `json:"description" db:"description" validate:"max=1000"`
	Price       float64   `json:"price" db:"price" validate:"required,gt=0"`
	Category    string    `json:"category" db:"category" validate:"max=100"`
	Stock       int       `json:"stock" db:"stock" validate:"gte=0"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description" validate:"max=1000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Category    string  `json:"category" validate:"max=100"`
	Stock       int     `json:"stock" validate:"gte=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=1,max=255"`
	Description string  `json:"description" validate:"omitempty,max=1000"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	Category    string  `json:"category" validate:"omitempty,max=100"`
	Stock       int     `json:"stock" validate:"omitempty,gte=0"`
}

type ProductResponse struct {
	Success bool     `json:"success"`
	Data    *Product `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type ProductsResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Data    []*Product `json:"data"`
	Total   int        `json:"total"`
}

var validate = validator.New()

func (p *CreateProductRequest) Validate() error {
	return validate.Struct(p)
}

func (p *UpdateProductRequest) Validate() error {
	return validate.Struct(p)
}
