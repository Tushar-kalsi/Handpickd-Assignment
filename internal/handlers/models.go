package handlers

import "github.com/google/uuid"

// ErrorResponse represents an error response
type ErrorResponse struct {
    Error string `json:"error"`
}

// ViewProductRequest represents a request to view a product
type ViewProductRequest struct {
    ProductID uuid.UUID `json:"product_id" binding:"required"`
}

// ProductResponse represents a product in the API response
type ProductResponse struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    ViewCount   int64     `json:"view_count"`
    CreatedAt   string    `json:"created_at,omitempty"`
    UpdatedAt   string    `json:"updated_at,omitempty"`
}

// TopProductsRequest represents a request to get top N products
type TopProductsRequest struct {
    Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
