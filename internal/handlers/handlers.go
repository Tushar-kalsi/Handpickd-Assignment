package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tushar-kalsi/product-views/internal/kafka"
	"github.com/tushar-kalsi/product-views/internal/repository"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	repo     repository.ProductRepository
	producer kafka.ProducerInterface
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(repo repository.ProductRepository, producer kafka.ProducerInterface) *ProductHandler {
	return &ProductHandler{
		repo:     repo,
		producer: producer,
	}
}

// ViewProduct handles the request to view a product
// @Summary Record a product view
// @Description Records a view for a specific product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param request body ViewProductRequest true "Product view request"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/view [post]
func (h *ProductHandler) ViewProduct(c *gin.Context) {
	var req ViewProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// Send view event to Kafka
	if err := h.producer.SendViewEvent(c.Request.Context(), req.ProductID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to record view"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "View recorded successfully",
	})
}

// GetTopProducts returns the top N most viewed products
// @Summary Get top N most viewed products
// @Description Returns the most viewed products, limited by the 'limit' parameter (max 100)
// @Tags products
// @Produce json
// @Param limit query int false "Maximum number of products to return (1-100)" default(10)
// @Success 200 {array} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products/top [get]
func (h *ProductHandler) GetTopProducts(c *gin.Context) {
	var req TopProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid query parameters"})
		return
	}

	// Default to 10 if limit is not provided
	if req.Limit == 0 {
		req.Limit = 10
	}

	products, err := h.repo.GetTopViewedProducts(c.Request.Context(), req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch top products"})
		return
	}

	// Convert repository models to API response models
	response := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		response = append(response, ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			ViewCount:   p.ViewCount,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct handles the request to get a product by ID
// @Summary Get a product by ID
// @Description Returns the product with the specified ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
		return
	}

	product, err := h.repo.GetProduct(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch product"})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		ViewCount:   product.ViewCount,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	})
}

// CreateProduct handles the request to create a new product
// @Summary Create a new product
// @Description Creates a new product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param request body ProductResponse true "Product details"
// @Success 201 {object} ProductResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req ProductResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	product := &repository.Product{
		Name:        req.Name,
		Description: req.Description,
		ViewCount:   req.ViewCount,
	}

	if err := h.repo.CreateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		ViewCount:   product.ViewCount,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	})
}
