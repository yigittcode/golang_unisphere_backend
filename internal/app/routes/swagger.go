package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/yigit/unisphere/docs"
)

// SetupSwagger configures Swagger documentation routes
func SetupSwagger(router *gin.Engine) {
	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
