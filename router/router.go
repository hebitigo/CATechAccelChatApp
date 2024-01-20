package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"CATechAccelChatApp/handler"
	"CATechAccelChatApp/repository"
	"CATechAccelChatApp/usecase"
)

func InitRouter(db *bun.DB, ctx context.Context) *gin.Engine {
	botEndpointRepository := repository.NewBotEndpointRepository(db, ctx)
	botEndpointUsecase := usecase.NewBotEndpointUsecase(botEndpointRepository)
	botEndpointHandler := handler.NewBotEndpointHandler(botEndpointUsecase)
	r := gin.Default()
	r.POST("/registerBotEndpoint", botEndpointHandler.RegisterBotEndpoint)
	return r
}
