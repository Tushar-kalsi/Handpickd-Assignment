package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tushar-kalsi/product-views/internal/repository"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) IncrementViewCount(ctx context.Context, productID uuid.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductRepository) GetTopViewedProducts(ctx context.Context, limit int) ([]repository.Product, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]repository.Product), args.Error(1)
}

func (m *MockProductRepository) GetProduct(ctx context.Context, id uuid.UUID) (*repository.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Product), args.Error(1)
}

func (m *MockProductRepository) CreateProduct(ctx context.Context, p *repository.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

// MockKafkaProducer is a mock implementation of Kafka Producer
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) SendViewEvent(ctx context.Context, productID uuid.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() {
	m.Called()
}

func TestViewProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		mockProducer := new(MockKafkaProducer)
		handler := &ProductHandler{
			repo:     mockRepo,
			producer: mockProducer,
		}

		productID := uuid.New()
		mockProducer.On("SendViewEvent", mock.Anything, productID).Return(nil)

		router := gin.New()
		router.POST("/view", handler.ViewProduct)

		reqBody := ViewProductRequest{ProductID: productID}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/view", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "View recorded successfully", response["message"])

		mockProducer.AssertExpectations(t)
	})

	t.Run("Invalid Product ID", func(t *testing.T) {
		handler := &ProductHandler{}

		router := gin.New()
		router.POST("/view", handler.ViewProduct)

		reqBody := map[string]string{"product_id": "invalid-uuid"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/view", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Kafka Producer Error", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		mockProducer := new(MockKafkaProducer)
		handler := &ProductHandler{
			repo:     mockRepo,
			producer: mockProducer,
		}

		productID := uuid.New()
		mockProducer.On("SendViewEvent", mock.Anything, productID).Return(errors.New("kafka error"))

		router := gin.New()
		router.POST("/view", handler.ViewProduct)

		reqBody := ViewProductRequest{ProductID: productID}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/view", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockProducer.AssertExpectations(t)
	})
}

func TestGetTopProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success with default limit", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := &ProductHandler{repo: mockRepo}

		products := []repository.Product{
			{
				ID:        uuid.New(),
				Name:      "Product 1",
				ViewCount: 1000,
			},
			{
				ID:        uuid.New(),
				Name:      "Product 2",
				ViewCount: 800,
			},
		}

		mockRepo.On("GetTopViewedProducts", mock.Anything, 10).Return(products, nil)

		router := gin.New()
		router.GET("/top", handler.GetTopProducts)

		req := httptest.NewRequest("GET", "/top", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []ProductResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "Product 1", response[0].Name)
		assert.Equal(t, int64(1000), response[0].ViewCount)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with custom limit", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := &ProductHandler{repo: mockRepo}

		products := make([]repository.Product, 50)
		for i := 0; i < 50; i++ {
			products[i] = repository.Product{
				ID:        uuid.New(),
				Name:      fmt.Sprintf("Product %d", i+1),
				ViewCount: int64(1000 - i*10),
			}
		}

		mockRepo.On("GetTopViewedProducts", mock.Anything, 50).Return(products, nil)

		router := gin.New()
		router.GET("/top", handler.GetTopProducts)

		req := httptest.NewRequest("GET", "/top?limit=50", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []ProductResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 50)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Limit exceeds maximum", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		mockProducer := new(MockKafkaProducer)
		handler := &ProductHandler{repo: mockRepo, producer: mockProducer}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/products/top?limit=150", nil)

		// No mock expectation needed as it should return 400 before calling repo

		handler.GetTopProducts(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database error", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := &ProductHandler{repo: mockRepo}

		mockRepo.On("GetTopViewedProducts", mock.Anything, 10).Return([]repository.Product{}, errors.New("db error"))

		router := gin.New()
		router.GET("/top", handler.GetTopProducts)

		req := httptest.NewRequest("GET", "/top", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
