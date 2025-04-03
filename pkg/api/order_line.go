package api

import (
	"context"
	"net/http"
	"postui_api/pkg/database"
	"postui_api/pkg/models"

	"github.com/gin-gonic/gin"
)

type OrderLineRepository interface {
	CreateOrderLine(c *gin.Context)
	FindOrderLine(c *gin.Context)
	UpdateOrderLine(c *gin.Context)
	DeleteOrderLine(c *gin.Context)
}

// orderLineRepository holds shared resources like database
type orderLineRepository struct {
	DB  database.Database
	Ctx *context.Context
}

// NewAppContext creates a new AppContext
func NewOrderLineRepository(db database.Database, ctx *context.Context) *orderLineRepository {
	return &orderLineRepository{
		DB:  db,
		Ctx: ctx,
	}
}

// @BasePath /api/v1

// CreateOrderLine godoc
// @Summary Create a new orderLine
// @Description Create a new orderLine with the given input data
// @Tags orderLines
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param   input     body   models.CreateOrderLine   true   "Create orderLine object"
// @Success 201 {object} models.OrderLine "Successfully created orderLine"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /order_lines [post]
func (r *orderLineRepository) CreateOrderLine(c *gin.Context) {
	appCtx, exists := c.MustGet("appCtxOrderLine").(*orderLineRepository)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var inputs []models.CreateOrderLine

	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var orderLines []models.OrderLine

	for _, input := range inputs {
		orderLine := models.OrderLine{ProductID: input.ProductID, Quantity: input.Quantity, Price: input.Price, Vat: input.Vat, Total: input.Total}
		orderLines = append(orderLines, orderLine)
	}

	appCtx.DB.Create(&orderLines)

	c.JSON(http.StatusCreated, gin.H{"data": orderLines})
}

// FindOrderLine godoc
// @Summary Find a orderLine by ID
// @Description Get details of a orderLine by its ID
// @Tags orderLines
// @Security JwtAuth
// @Produce json
// @Param id path string true "OrderLine ID"
// @Success 200 {object} models.OrderLine "Successfully retrieved orderLine"
// @Failure 404 {string} string "OrderLine not found"
// @Router /order_lines/{id} [get]
func (r *orderLineRepository) FindOrderLine(c *gin.Context) {
	var orderLine models.OrderLine

	if err := r.DB.Where("id = ?", c.Param("id")).First(&orderLine).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "orderLine not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orderLine})
}

// UpdateOrderLine godoc
// @Summary Update a orderLine by ID
// @Description Update the orderLine details for the given ID
// @Tags orderLines
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param id path string true "OrderLine ID"
// @Param input body models.UpdateOrderLine true "Update orderLine object"
// @Success 200 {object} models.OrderLine "Successfully updated orderLine"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "orderLine not found"
// @Router /order_lines/{id} [put]
func (r *orderLineRepository) UpdateOrderLine(c *gin.Context) {
	var orderLine models.OrderLine
	var input models.UpdateOrderLine

	if err := r.DB.Where("id = ?", c.Param("id")).First(&orderLine).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "orderLine not found"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.DB.Model(&orderLine).Updates(models.OrderLine{ProductID: input.ProductID, Quantity: input.Quantity, Price: input.Price, Vat: input.Vat, Total: input.Total})

	c.JSON(http.StatusOK, gin.H{"data": orderLine})
}

// DeleteOrderLine godoc
// @Summary Delete a orderLine by ID
// @Description Delete the orderLine with the given ID
// @Tags orderLines
// @Security JwtAuth
// @Produce json
// @Param id path string true "OrderLine ID"
// @Success 204 {string} string "Successfully deleted orderLine"
// @Failure 404 {string} string "orderLine not found"
// @Router /order_lines/{id} [delete]
func (r *orderLineRepository) DeleteOrderLine(c *gin.Context) {
	var orderLine models.OrderLine

	if err := r.DB.Where("id = ?", c.Param("id")).First(&orderLine).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "orderLine not found"})
		return
	}

	r.DB.Delete(&orderLine)

	c.JSON(http.StatusNoContent, gin.H{"data": true})
}
