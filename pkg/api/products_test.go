package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"postui_api/pkg/cache"
	"postui_api/pkg/database"
	"postui_api/pkg/models"
	"testing"

	"gorm.io/gorm"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewProductRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	mockCtx := context.Background()

	repo := NewProductRepository(mockDB, mockCache, &mockCtx)

	assert.NotNil(t, repo, "NewProductRepository should return a non-nil instance of productRepository")
	assert.Equal(t, mockDB, repo.DB, "DB should be set to the mock database instance")
	assert.Equal(t, mockCache, repo.RedisClient, "RedisClient should be set to the mock cache instance")
}

func TestHealthcheck(t *testing.T) {
	// Set up the mock controller and the mocked dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Set up the Gin context with a response recorder
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	_, router := gin.CreateTestContext(recorder)

	// Create a mock repository and expect the Healthcheck method to be called
	mockRepo := NewMockProductRepository(ctrl)
	mockRepo.EXPECT().Healthcheck(gomock.Any()).Do(func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok") // Explicitly setting the response here
	})

	// Setting up a basic GET route to test Healthcheck
	router.GET("/healthcheck", mockRepo.Healthcheck)

	// Perform the GET request
	req, _ := http.NewRequest(http.MethodGet, "/healthcheck", nil)
	router.ServeHTTP(recorder, req)

	// Check the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "\"ok\"", recorder.Body.String())
}

func TestFindProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	mockGormDB := database.NewMockDatabase(ctrl) // Correct type for GORM DB operations
	ctx := context.Background()

	repo := NewProductRepository(mockDB, mockCache, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/products", repo.FindProducts)

	// Set up common mock expectations
	stock, _ := decimal.NewFromString("100")
	mockGormDB.EXPECT().Find(gomock.Any()).DoAndReturn(func(products *[]models.Product) *gorm.DB {
		*products = append(*products, models.Product{Name: "New Product", Price: 10, Vat: 2100, Stock: stock, BarcodeNumber: "465677261626"})
		return &gorm.DB{Error: nil} // Assume this is the struct provided by the actual Gorm package
	}).AnyTimes()

	products := []models.Product{{Name: "Product One", Price: 10, Vat: 2100, Stock: stock, BarcodeNumber: "465677261626"}}
	cachedData, _ := json.Marshal(products)
	mockCache.EXPECT().Get(ctx, "products_offset_0_limit_10").Return(redis.NewStringResult(string(cachedData), nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/products?offset=0&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Product One")
}

func TestCreateProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	mockCache := cache.NewMockCache(ctrl)
	ctx := context.Background()

	repo := NewProductRepository(mockDB, mockCache, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/products", func(c *gin.Context) {
		// Set the appCtx in the Gin context
		c.Set("appCtxProduct", repo)
		repo.CreateProducts(c)
	})

	// Example data for the test
	stock, _ := decimal.NewFromString("100")
	inputProducts := []models.CreateProducts{
		{Name: "New Product", Price: 10, Vat: 2100, Stock: stock, BarcodeNumber: "465677261626"},
		{Name: "Another Product", Price: 20, Vat: 1900, Stock: stock, BarcodeNumber: "461246179231"},
	}
	requestBody, err := json.Marshal(inputProducts)
	if err != nil {
		t.Fatalf("Failed to marshal input product data: %v", err)
	}

	// Set up database mock to simulate successful product creation
	mockDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(product *[]models.Product) *gorm.DB {
		// Normally, you might simulate setting an ID or other fields modified by the DB
		return &gorm.DB{Error: nil}
	})

	// Set up cache mock to simulate key retrieval and deletion
	keyPattern := "products_offset_*"
	mockCache.EXPECT().Keys(ctx, keyPattern).Return(redis.NewStringSliceResult([]string{"products_offset_0_limit_10"}, nil))
	mockCache.EXPECT().Del(ctx, "products_offset_0_limit_10").Return(redis.NewIntResult(1, nil))

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/products", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create the HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Serve the HTTP request
	r.ServeHTTP(w, req)

	// Assertions to check the response
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status code 201")
	assert.Contains(t, w.Body.String(), "New Product", "Response body should contain the product title")
}

func TestFindProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewProductRepository(mockDB, nil, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/product/:id", repo.FindProduct)

	// Prepare test data
	stock, _ := decimal.NewFromString("100")
	expectedProduct := models.Product{
		ID:            1,
		Name:          "My Product",
		Price:         10,
		Stock:         stock,
		Vat:           2100,
		BarcodeNumber: "465677261626",
	}

	// Mock expectations

	// Mock the Where method
	mockDB.EXPECT().
		Where("id = ?", "1").
		DoAndReturn(func(query interface{}, args ...interface{}) database.Database {
			// Return mockDB to allow method chaining
			return mockDB
		}).Times(1)

	// Mock the First method
	mockDB.EXPECT().
		First(gomock.Any()).
		DoAndReturn(func(dest interface{}, conds ...interface{}) database.Database {
			if b, ok := dest.(*models.Product); ok {
				*b = expectedProduct
			}
			return mockDB
		}).Times(1)

	// Mock the Error method or field access
	mockDB.EXPECT().
		Error().
		Return(nil).
		Times(1)

	// Perform the request
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/product/1", nil)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Status  int            `json:"status"`
		Message string         `json:"message"`
		Data    models.Product `json:"data"`
	}

	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedProduct.ID, response.Data.ID)
	assert.Equal(t, expectedProduct.Name, response.Data.Name)
	assert.Equal(t, expectedProduct.Price, response.Data.Price)
	assert.Equal(t, expectedProduct.Stock, response.Data.Stock)
	assert.Equal(t, expectedProduct.Vat, response.Data.Vat)
	assert.Equal(t, expectedProduct.BarcodeNumber, response.Data.BarcodeNumber)
}

func TestDeleteProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock for the database
	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewProductRepository(mockDB, nil, &ctx)

	// Set up Gin for testing
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/product/:id", repo.DeleteProduct)

	// Prepare the product data
	stock, _ := decimal.NewFromString("100")
	existingProduct := models.Product{
		ID:            1,
		Name:          "My Product",
		Price:         10,
		Stock:         stock,
		Vat:           2100,
		BarcodeNumber: "465677261626",
	}

	// Mock Where to return the existingProduct for chaining
	mockDB.EXPECT().
		Where("id = ?", "1").
		Return(mockDB).Times(1)

	// Mock First to load the existingProduct and return mockDB
	mockDB.EXPECT().
		First(gomock.Any()).
		DoAndReturn(func(dest interface{}, conds ...interface{}) database.Database {
			if b, ok := dest.(*models.Product); ok {
				*b = existingProduct
			}
			return mockDB
		}).Times(1)

	// Mock Delete method
	mockDB.EXPECT().
		Delete(&existingProduct).
		Return(&gorm.DB{Error: nil}).Times(1)

	// Mock Error method to return nil
	mockDB.EXPECT().Error().Return(nil).AnyTimes()

	// Perform the DELETE request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/product/1", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNoContent, w.Code)
}
