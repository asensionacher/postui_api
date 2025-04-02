package api

import (
	"context"
	"postui_api/pkg/cache"
	"postui_api/pkg/database"
	"postui_api/pkg/middleware"
	"time"

	docs "postui_api/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"golang.org/x/time/rate"
)

func ContextMiddleware(productRepository ProductRepository, orderRepository OrderRepository, orderLineRepository OrderLineRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("appCtxProduct", productRepository)
		c.Next()
		c.Set("appCtxOrder", orderRepository)
		c.Next()
		c.Set("appCtxOrderLine", orderLineRepository)
		c.Next()
	}
}

func NewRouter(logger *zap.Logger, mongoCollection *mongo.Collection, db database.Database, redisClient cache.Cache, ctx *context.Context) *gin.Engine {
	productRepository := NewProductRepository(db, redisClient, ctx)
	userRepository := NewUserRepository(db, ctx)
	orderLineRepository := NewOrderLineRepository(db, ctx)
	orderRepository := NewOrderRepository(db, ctx)

	r := gin.Default()
	r.Use(ContextMiddleware(productRepository, orderRepository, orderLineRepository))

	//r.Use(gin.Logger())
	r.Use(middleware.Logger(logger, mongoCollection))
	if gin.Mode() == gin.ReleaseMode {
		r.Use(middleware.Security())
		r.Use(middleware.Xss())
	}
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimiter(rate.Every(1*time.Minute), 60)) // 60 requests per minute

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", productRepository.Healthcheck)                                                              // No need to be admin
		v1.GET("/products", middleware.JWTAuth(), productRepository.FindProducts)                               // No need to be admin
		v1.POST("/products", middleware.JWTAuth(), middleware.IsAdmin(), productRepository.CreateProducts)      // Need to be admin
		v1.GET("/products/:id", middleware.JWTAuth(), productRepository.FindProduct)                            // No need to be admin
		v1.PUT("/products/:id", middleware.JWTAuth(), middleware.IsAdmin(), productRepository.UpdateProduct)    // Need to be admin
		v1.DELETE("/products/:id", middleware.JWTAuth(), middleware.IsAdmin(), productRepository.DeleteProduct) // Need to be admin
		v1.POST("/order_lines", middleware.JWTAuth(), orderLineRepository.CreateOrderLine)                      // No need to be admin
		v1.GET("/order_lines/:id", middleware.JWTAuth(), orderLineRepository.FindOrderLine)                     // No need to be admin
		v1.PUT("/order_lines/:id", middleware.JWTAuth(), orderLineRepository.UpdateOrderLine)                   // No need to be admin
		v1.DELETE("/order_lines/:id", middleware.JWTAuth(), orderLineRepository.DeleteOrderLine)                // No need to be admin
		v1.POST("/orders", middleware.JWTAuth(), orderRepository.CreateOrder)                                   // No need to be admin
		v1.GET("/orders/:id", middleware.JWTAuth(), orderRepository.FindOrder)                                  // No need to be admin
		v1.PUT("/orders/:id", middleware.JWTAuth(), orderRepository.UpdateOrder)                                // No need to be admin
		v1.DELETE("/orders/:id", middleware.JWTAuth(), orderRepository.DeleteOrder)                             // No need to be admin

		v1.POST("/login", userRepository.LoginHandler)                                                             // No need to be admin neither to be logged
		v1.POST("/register", middleware.JWTAuth(), middleware.IsAdmin(), userRepository.RegisterHandler)           // Need to be admin
		v1.POST("/resetPassword", middleware.JWTAuth(), middleware.IsAdmin(), userRepository.ResetPasswordHandler) // Need to be admin
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
