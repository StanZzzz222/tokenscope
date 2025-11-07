package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"tokenscope/api/handler"
	"tokenscope/api/middleware"
	"tokenscope/common/logger"
	"tokenscope/models"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: router.go
*/

func Run() {
	if !viper.GetBool("server.debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.CORSMiddleware())
	r.NoRoute(func(ctx *gin.Context) {
		models.Response(ctx).Error(http.StatusNotFound, fmt.Errorf("%v does not exist", ctx.Request.RequestURI))
	})
	apiV1 := r.Group("/api_v1")
	{
		handler.InitAssetRouters(apiV1)
		handler.InitBlockchainRouters(apiV1)
	}
	logger.Logger().Infof("server start, url: http://localhost:%s", viper.GetString("server.port"))
	go func() {
		port := viper.GetString("server.port")
		err := r.Run(fmt.Sprintf(":%s", port))
		if err != nil {
			panic(fmt.Sprintf("Server start error, port: %v is already in use by another program", port))
			return
		}
	}()
}
