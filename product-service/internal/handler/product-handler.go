package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/product-service/internal/model"
	"github.com/leandrowiemesfilho/product-service/internal/service"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service service.ProductService
	logger  *zap.SugaredLogger
}

func NewProductHandler(service service.ProductService, logger *zap.SugaredLogger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnw("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, model.ProductResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	product, err := h.service.CreateProduct(&req)
	if err != nil {
		h.logger.Errorw("Failed to create product", "error", err)
		c.JSON(http.StatusInternalServerError, model.ProductResponse{
			Success: false,
			Error:   "Failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, model.ProductResponse{
		Success: true,
		Data:    product,
	})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.ProductResponse{
			Success: false,
			Error:   "Product ID is required",
		})
		return
	}

	product, err := h.service.GetProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, model.ProductResponse{
				Success: false,
				Error:   "Product not found",
			})
			return
		}

		h.logger.Errorw("Failed to get product", "error", err, "product_id", id)
		c.JSON(http.StatusInternalServerError, model.ProductResponse{
			Success: false,
			Error:   "Failed to get product",
		})
		return
	}

	c.JSON(http.StatusOK, model.ProductResponse{
		Success: true,
		Data:    product,
	})
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.service.GetAllProducts()
	if err != nil {
		h.logger.Errorw("Failed to get products", "error", err)
		c.JSON(http.StatusInternalServerError, model.ProductsResponse{
			Success: false,
			Error:   "Failed to get products",
		})
		return
	}

	c.JSON(http.StatusOK, model.ProductsResponse{
		Success: true,
		Data:    products,
		Total:   len(products),
	})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.ProductResponse{
			Success: false,
			Error:   "Product ID is required",
		})
		return
	}

	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnw("Invalid request body", "error", err, "product_id", id)
		c.JSON(http.StatusBadRequest, model.ProductResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	product, err := h.service.UpdateProduct(id, &req)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, model.ProductResponse{
				Success: false,
				Error:   "Product not found",
			})
			return
		}

		h.logger.Errorw("Failed to update product", "error", err, "product_id", id)
		c.JSON(http.StatusInternalServerError, model.ProductResponse{
			Success: false,
			Error:   "Failed to update product",
		})
		return
	}

	c.JSON(http.StatusOK, model.ProductResponse{
		Success: true,
		Data:    product,
	})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	err := h.service.DeleteProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		h.logger.Errorw("Failed to delete product", "error", err, "product_id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
