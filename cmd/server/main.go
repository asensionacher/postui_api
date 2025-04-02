package main

import (
	"context"
	"log"
	"postui_api/pkg/api"
	"postui_api/pkg/cache"
	"postui_api/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

//	@title			Swagger POS TUI API
//	@version		1.0
//	@description	This is the API server used for POS TUI.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-git clone TODO: CHANGE

//	@host		localhost:8001
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	JwtAuth
//	@in							header
//	@name						Authorization

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	redisClient := cache.NewRedisClient()
	db := database.NewDatabase()
	dbWrapper := &database.GormDatabase{DB: db}
	mongo := database.SetupMongoDB()
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)

	r := api.NewRouter(logger, mongo, dbWrapper, redisClient, &ctx)

	if err := r.Run(":8001"); err != nil {
		log.Fatal(err)
	}
}
