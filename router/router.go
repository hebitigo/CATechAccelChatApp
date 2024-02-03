package router

import (
	"context"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"github.com/hebitigo/CATechAccelChatApp/handler"
	"github.com/hebitigo/CATechAccelChatApp/repository"
	"github.com/hebitigo/CATechAccelChatApp/usecase"
	"github.com/hebitigo/CATechAccelChatApp/ws"
)

func InitRouter(db *bun.DB, ctx context.Context) *gin.Engine {

	//TODO:https://github.com/code-kakitai/code-kakitai/blob/main/app/server/route/route.go#L79
	//を参考にして、handler毎に分けてrouteを初期化する
	botEndpointRepository := repository.NewBotEndpointRepository(db, ctx)
	botEndpointUsecase := usecase.NewBotEndpointUsecase(botEndpointRepository)
	botEndpointHandler := handler.NewBotEndpointHandler(botEndpointUsecase)
	r := gin.Default()
	//TODO:https://github.com/code-kakitai/code-kakitai/blob/main/app/presentation/settings/gin.go#L10
	//を参考にして*gin.Engineにcorsの設定を追加する
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))
	r.POST("/bot_endpoint", botEndpointHandler.RegisterBotEndpoint)

	serverRepository := repository.NewServerRepository(db)
	channelRepository := repository.NewChannelRepository(db)
	userServerRepository := repository.NewUserServerRepository(db)
	txRepository := repository.NewTxRepository(db)
	serverUsecase := usecase.NewServerUsecase(serverRepository, channelRepository, userServerRepository, txRepository)
	serverHandler := handler.NewServerHandler(serverUsecase)
	r.POST("/server", serverHandler.RegisterServer)

	r.GET("/servers/:user_Id", serverHandler.GetServersByUserID)

	userRepostiory := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepostiory)
	userHandler := handler.NewUserHandler(userUsecase)
	r.POST("/user/upsert", userHandler.UpsertUser)

	hub := ws.NewHub()
	go hub.Run()
	messageRepository := repository.NewMessageRepository(db)
	wsHandler := ws.NewHandler(hub, messageRepository)
	r.GET("/ws/channel/join/:server_Id/:channel_Id/:user_Id", wsHandler.JoinChannel)

	return r
}
