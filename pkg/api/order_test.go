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
	"github.com/lib/pq"


	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	mockCtx := context.Background()

	repo := NewOrderRepository(mockDB, &mockCtx)

	assert.NotNil(t, repo, "NewOrderRepository should return a non-nil instance of orderRepository")
	assert.Equal(t, mockDB, repo.DB, "DB should be set to the mock database instance")
}

func TestCreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()

	repo := NewOrderRepository(mockDB, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/orders", func(c *gin.Context) {
		// Set the appCtx in the Gin context
		c.Set("appCtxOrder", repo)
		repo.CreateOrder(c)
	})

	// Example data for the test
	quantity, _ := decimal.NewFromString("100")

	line1 := models.OrderLine{
		ID:        1,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}

	line2 := models.OrderLine{
		ID:        2,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}
	var lines_id pq.Int64Array
	lines_id = append(lines_id, line1.ID)
	lines_id = append(lines_id, line2.ID)

	inputOrders := models.CreateOrder{Vendor: "username", Total: 1000, LinesID: lines_id, CashoutNumber: 1}

	requestBody, err := json.Marshal(inputOrders)
	if err != nil {
		t.Fatalf("Failed to marshal input order data: %v", err)
	}

	// Set up database mock to simulate successful order creation
	mockDB.EXPECT().Create(gomock.Any()).DoAndReturn(func(order *models.Order) *gorm.DB {
		// Normally, you might simulate setting an ID or other fields modified by the DB
		return &gorm.DB{Error: nil}
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/orders", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to create the HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Serve the HTTP request
	r.ServeHTTP(w, req)

	// Assertions to check the response
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status code 201")
	assert.Contains(t, w.Body.String(), "username", "Response body should contain the order Vat")
}

func TestFindOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewOrderRepository(mockDB, &ctx)

	// Set up Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/order/:id", repo.FindOrder)

	// Prepare test data
	quantity, _ := decimal.NewFromString("100")

	line1 := models.OrderLine{
		ID:        1,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}

	line2 := models.OrderLine{
		ID:        2,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}
	var lines_id pq.Int64Array
	lines_id = append(lines_id, line1.ID)
	lines_id = append(lines_id, line2.ID)

	expectedOrder := models.Order{
		ID:            1,
		Vendor:        "username",
		Total:         1000,
		LinesID:       lines_id,
		CashoutNumber: 1,
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
			if b, ok := dest.(*models.Order); ok {
				*b = expectedOrder
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
	req := httptest.NewRequest(http.MethodGet, "/order/1", nil)
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Status  int          `json:"status"`
		Message string       `json:"message"`
		Data    models.Order `json:"data"`
	}

	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.ID, response.Data.ID)
	assert.Equal(t, expectedOrder.Vendor, response.Data.Vendor)
	assert.Equal(t, expectedOrder.Total, response.Data.Total)
	assert.Equal(t, expectedOrder.LinesID, response.Data.LinesID)
	assert.Equal(t, expectedOrder.CashoutNumber, response.Data.CashoutNumber)
}

func TestDeleteOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock for the database
	mockDB := database.NewMockDatabase(ctrl)
	ctx := context.Background()
	repo := NewOrderRepository(mockDB, &ctx)

	// Set up Gin for testing
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/order/:id", repo.DeleteOrder)

	// Prepare the order data
	quantity, _ := decimal.NewFromString("100")

	line1 := models.OrderLine{
		ID:        1,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}

	line2 := models.OrderLine{
		ID:        2,
		ProductID: 1,
		Quantity:  quantity,
		Price:     100,
		Vat:       2100,
		Total:     121,
	}
	var lines_id pq.Int64Array
	lines_id = append(lines_id, line1.ID)
	lines_id = append(lines_id, line2.ID)

	existingOrder := models.Order{
		ID:            1,
		Vendor:        "username",
		Total:         1000,
		LinesID:       lines_id,
		CashoutNumber: 1,
	}

	// Mock Where to return the existingOrder for chaining
	mockDB.EXPECT().
		Where("id = ?", "1").
		Return(mockDB).Times(1)

	// Mock First to load the existingOrder and return mockDB
	mockDB.EXPECT().
		First(gomock.Any()).
		DoAndReturn(func(dest interface{}, conds ...interface{}) database.Database {
			if b, ok := dest.(*models.Order); ok {
				*b = existingOrder
			}
			return mockDB
		}).Times(1)

	// Mock Delete method
	mockDB.EXPECT().
		Delete(&existingOrder).
		Return(&gorm.DB{Error: nil}).Times(1)

	// Mock Error method to return nil
	mockDB.EXPECT().Error().Return(nil).AnyTimes()

	// Perform the DELETE request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/order/1", nil)
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusNoContent, w.Code)
}
