package api

import (
	"context"
	"encoding/json"
	"net/http"
	"postui_api/pkg/cache"
	"postui_api/pkg/database"
	"postui_api/pkg/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductRepository interface {
	Healthcheck(c *gin.Context)
	FindProducts(c *gin.Context)
	CreateProducts(c *gin.Context)
	FindProduct(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
}

// productRepository holds shared resources like database and Redis client
type productRepository struct {
	DB          database.Database
	RedisClient cache.Cache
	Ctx         *context.Context
}

// NewAppContext creates a new AppContext
func NewProductRepository(db database.Database, redisClient cache.Cache, ctx *context.Context) *productRepository {
	return &productRepository{
		DB:          db,
		RedisClient: redisClient,
		Ctx:         ctx,
	}
}

// @BasePath /api/v1

// Healthcheck godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} ok
// @Router / [get]
func (r *productRepository) Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

// FindProducts godoclet response = client
            .post("http://localhost:8001/api/v1/orders")
            .header("Authorization", format!("Bearer {}", token))
            .json(&order)
            .send()
            .await?;

        match response.status() {
            StatusCode::CREATED => {
                *state.cart.lock().await = Vec::new();
                Ok(())
            }
            status => {
                let body = response.text().await?;
                Err(anyhow::anyhow!("Checkout failed ({}): {}", status, body))
            }
        }
// @Security JwtAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for paginaCreateProducttion" default(10)
// @Success 200 {array} models.Product "Successfully retrieved list of products"
// @Router /products [get]
func (r *productRepository) FindProducts(c *gin.Context) {
	var products []models.Product

	// Get query params
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	// Convert query params to integers
	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset format"})
		return
	}
	var total_items int64

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit format"})
		return
	}

	r.DB.Model(&models.Product{}).Count(&total_items)
	total_pages := total_items / int64(limit)

	// Create a cache key based on query params
	cacheKey := "products_offset_" + offsetQuery + "_limit_" + limitQuery
	// Try fetching the data from Redis first
	cachedProducts, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(cachedProducts), &products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal cached data"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": products,
			"pagination": gin.H{
				"total_items": total_items,
				"page":        offset,
				"limit":       limit,
				"total_pages": total_pages,
			},
		})
		return
	}

	// If cache missed, fetch data from the database with proper pagination
	result := r.DB.Offset(offset).Limit(limit).Find(&products)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// Serialize products object and store it in Redis
	serializedProducts, err := json.Marshal(products)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal data"})
		return
	}

	// Set cache with expiration time
	err = r.RedisClient.Set(*r.Ctx, cacheKey, serializedProducts, time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"pagination": gin.H{
			"total_items": total_items,
			"page":        offset,
			"limit":       limit,
			"total_pages": total_pages,
		},
	})
}

// CreateProducts godoc
// @Summary Create new products
// @Description Create new products with the given input data
// @Tags products
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param   input     body   []models.CreateProducts   true   "Create product object"
// @Success 201 {object} []models.Product "Successfully created product"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /products [post]
func (r *productRepository) CreateProducts(c *gin.Context) {
	appCtx, exists := c.MustGet("appCtxProduct").(*productRepository)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var inputs []models.CreateProducts

	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var products []models.Product
	for _, input := range inputs {
		product := models.Product{Name: input.Name, Price: input.Price, Vat: input.Vat, Stock: input.Stock, BarcodeNumber: input.BarcodeNumber}
		products = append(products, product)
	}

	appCtx.DB.Create(&products)

	// Invalidate cache
	keysPattern := "products_offset_*"
	keys, err := appCtx.RedisClient.Keys(*appCtx.Ctx, keysPattern).Result()
	if err == nil {
		for _, key := range keys {
			appCtx.RedisClient.Del(*appCtx.Ctx, key)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"data": products})
}

// FindProduct godoc
// @Summary Find a product by ID
// @Description Get details of a product by its ID
// @Tags products
// @Security JwtAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product "Successfully retrieved product"
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [get]
func (r *productRepository) FindProduct(c *gin.Context) {
	var product models.Product

	if err := r.DB.Where("id = ?", c.Param("id")).First(&product).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// UpdateProduct godoc
// @Summary Update a product by ID
// @Description Update the product details for the given ID
// @Tags products
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param id path string true "Product ID"
// @Param input body models.UpdateProduct true "Update product object"
// @Success 200 {object} models.Product "Successfully updated product"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "product not found"
// @Router /products/{id} [put]
func (r *productRepository) UpdateProduct(c *gin.Context) {
	var product models.Product
	var input models.UpdateProduct

	if err := r.DB.Where("id = ?", c.Param("id")).First(&product).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.DB.Model(&product).Updates(models.Product{Name: input.Name, Price: input.Price, Vat: input.Vat, Stock: input.Stock, BarcodeNumber: input.BarcodeNumber})

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// DeleteProduct godoc
// @Summary Delete a product by ID
// @Description Delete the product with the given ID
// @Tags products
// @Security JwtAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 {string} string "Successfully deleted product"
// @Failure 404 {string} string "product not found"
// @Router /products/{id} [delete]
func (r *productRepository) DeleteProduct(c *gin.Context) {
	var product models.Product

	if err := r.DB.Where("id = ?", c.Param("id")).First(&product).Error(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	r.DB.Delete(&product)

	c.JSON(http.StatusNoContent, gin.H{"data": true})
}
