package api

import (
	"context"
	"net/http"
	"postui_api/pkg/database"
	"postui_api/pkg/models"

	"github.com/gin-gonic/gin"
)

type OrderRepository interface {
	CreateOrder(c *gin.Context)
	FindOrder(c *gin.Context)
	UpdateOrder(c *gin.Context)
	DeleteOrder(c *gin.Context)
}

// orderRepository holds shared resources like database
type orderRepository struct {
	DB  database.Database
	Ctx *context.Context
}

// NewAppContext creates a new AppContext
func NewOrderRepository(db database.Database, ctx *context.Context) *orderRepository {
	return &orderRepository{
		DB:  db,
		Ctx: ctx,
	}
}

// @BasePath /api/v1

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with the given input data
// @Tags orders
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param   input     body   models.CreateOrder   true   "Create order object"
// @Success 201 {object} models.Order "Successfully created order"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /order_lines [post]
func (r *orderRepository) CreateOrder(c *gin.Context) {
	appCtx, exists := c.MustGet("appCtxOrder").(*orderRepository)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var input models.CreateOrder

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := models.Order{Vendor: input.Vendor, Total: input.Total, LinesID: input.LinesID, CashoutNumber: input.CashoutNumber}

	appCtx.DB.Create(&order)

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

// FindOrder godoc
// @Summary Find an order by ID
// @Description Get details of an order by its ID
// @Tags orders
// @Security JwtAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.Order "Successfully retrieved order"
// @Failure 404 {string} string "Order not found"
// @Router /order_lines/{id} [get]
func (r *orderRepository) FindOrder(c *gin.Context) {
	var order models.Order

	if err := r.DB.Where("id = ?", c.Param("id")).First(&order).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// UpdateOrder godoc
// @Summary Update an order by ID
// @Description Update the order details for the given ID
// @Tags orders
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param id path string true "Order ID"
// @Param input body models.UpdateOrder true "Update order object"
// @Success 200 {object} models.Order "Successfully updated order"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "order not found"
// @Router /order_lines/{id} [put]
func (r *orderRepository) UpdateOrder(c *gin.Context) {
	var order models.Order
	var input models.UpdateOrder

	if err := r.DB.Where("id = ?", c.Param("id")).First(&order).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.DB.Model(&order).Updates(models.Order{Vendor: input.Vendor, Total: input.Total, LinesID: input.LinesID, CashoutNumber: input.CashoutNumber})

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// DeleteOrder godoc
// @Summary Delete an order by ID
// @Description Delete the order with the given ID
// @Tags orders
// @Security JwtAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 204 {string} string "Successfully deleted order"
// @Failure 404 {string} string "order not found"
// @Router /order_lines/{id} [delete]
func (r *orderRepository) DeleteOrder(c *gin.Context) {
	var order models.Order

	if err := r.DB.Where("id = ?", c.Param("id")).First(&order).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	r.DB.Delete(&order)

	c.JSON(http.StatusNoContent, gin.H{"data": true})
}
