package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"postui_api/pkg/database"
	"postui_api/pkg/models"
	"testing"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderLineRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	mockCtx := context.Background()

	repo := NewOrderLineRepository(mockDB, &mockCtx)

	assert.NotNil(t, repo, "NewOrderLineRepository should return a non-nil instance of orderLineRepository")
	assert.Equal(t, mockDB, repo.DB, "DB should be set to the mock database instance")
}

func TestCreateOrderLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()

	repo := NewOrderLineRepository(mockDB, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/order_lines", func(c *gin.Context) {
		// Set the appCtx in the Gin context
		c.Set("appCtxOrderLine", repo)
		repo.CreateOrderLine(c)
	})

	// Example data for the test
	quantity, _ := decimal.NewFromString("100")
	inputOrderLines := []models.CreateOrderLine{
		{ProductID: 1, Quantity: quantity, Price: 100, Vat: 2100, Total: 2},
	}

	requestBody, err := json.Marshal(inputOrderLines)
	if err != nil {
		t.Fatalf("Failed to marshal input orderLine data: %v", err)
	}

	// Set up database mock to simulate successful orderLine creation
	mockDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(orderLine *[]models.OrderLine) *gorm.DB {
		// Normally, you might simulate setting an ID or other fields modified by the DB
		return &gorm.DB{Error: nil}
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/order_lines", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create the HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Serve the HTTP request
	r.ServeHTTP(w, req)

	// Assertions to check the response
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status code 201")
	assert.Contains(t, w.Body.String(), "2100", "Response body should contain the orderLine Vat")
}

func TestFindOrderLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewOrderLineRepository(mockDB, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/orderLine/:id", repo.FindOrderLine)

	// Prepare test data
	quantity, _ := decimal.NewFromString("100")
	expectedOrderLine := models.OrderLine{
		ID:        1,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
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
			if b, ok := dest.(*models.OrderLine); ok {
				*b = expectedOrderLine
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
	req := httptest.NewRequest(http.MethodGet, "/orderLine/1", nil)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Status  int              `json:"status"`
		Message string           `json:"message"`
		Data    models.OrderLine `json:"data"`
	}

	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrderLine.ID, response.Data.ID)
	assert.Equal(t, expectedOrderLine.ProductID, response.Data.ProductID)
	assert.Equal(t, expectedOrderLine.Quantity, response.Data.Quantity)
	assert.Equal(t, expectedOrderLine.Price, response.Data.Price)
	assert.Equal(t, expectedOrderLine.Vat, response.Data.Vat)
	assert.Equal(t, expectedOrderLine.Total, response.Data.Total)
}

func TestDeleteOrderLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock for the database
	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewOrderLineRepository(mockDB, &ctx)

	// Set up Gin for testing
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/orderLine/:id", repo.DeleteOrderLine)

	// Prepare the orderLine data
	quantity, _ := decimal.NewFromString("100")
	existingOrderLine := models.OrderLine{
		ID:        1,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}

	// Mock Where to return the existingOrderLine for chaining
	mockDB.EXPECT().
		Where("id = ?", "1").
		Return(mockDB).Times(1)

	// Mock First to load the existingOrderLine and return mockDB
	mockDB.EXPECT().
		First(gomock.Any()).
		DoAndReturn(func(dest interface{}, conds ...interface{}) database.Database {
			if b, ok := dest.(*models.OrderLine); ok {
				*b = existingOrderLine
			}
			return mockDB
		}).Times(1)

	// Mock Delete method
	mockDB.EXPECT().
		Delete(&existingOrderLine).
		Return(&gorm.DB{Error: nil}).Times(1)

	// Mock Error method to return nil
	mockDB.EXPECT().Error().Return(nil).AnyTimes()

	// Perform the DELETE request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/orderLine/1", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNoContent, w.Code)
}
