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
	r := gin.Default()
	//TODO:https://github.com/code-kakitai/code-kakitai/blob/main/app/presentation/settings/gin.go#L10
	//を参考にして*gin.Engineにcorsの設定を追加する
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))

	//TODO:https://github.com/code-kakitai/code-kakitai/blob/main/app/server/route/route.go#L79
	//を参考にして、handler毎に分けてrouteを初期化する
	botEndpointRepository := repository.NewBotEndpointRepository(db, ctx)
	botEndpointUsecase := usecase.NewBotEndpointUsecase(botEndpointRepository)
	botEndpointHandler := handler.NewBotEndpointHandler(botEndpointUsecase)
	r.POST("/bot_endpoint", botEndpointHandler.RegisterBotEndpoint)

	serverRepository := repository.NewServerRepository(db)
	channelRepository := repository.NewChannelRepository(db)
	userServerRepository := repository.NewUserServerRepository(db)
	txRepository := repository.NewTxRepository(db)
	userRepostiory := repository.NewUserRepository(db)
	serverUsecase := usecase.NewServerUsecase(serverRepository, channelRepository, userServerRepository, txRepository, userRepostiory)
	serverHandler := handler.NewServerHandler(serverUsecase)
	r.POST("/server", serverHandler.RegisterServer)
	r.POST("/server/create/invitation", serverHandler.CreateInvitationByJWT)
	r.POST("/server/join", serverHandler.JoinServerByInvitation)
	r.GET("/servers/:user_id", serverHandler.GetServersByUserID)

	userUsecase := usecase.NewUserUsecase(userRepostiory)
	userHandler := handler.NewUserHandler(userUsecase)
	r.POST("/user/upsert", userHandler.UpsertUser)

	channelUsecase := usecase.NewChannelUsecase(channelRepository)
	channelHandler := handler.NewChannelHandler(channelUsecase)
	r.POST("/channel", channelHandler.RegisterChannel)
	r.GET("/channels/:server_id", channelHandler.GetChannelsByServerID)

	hub := ws.NewHub()
	go hub.Run()
	messageRepository := repository.NewMessageRepository(db)
	wsHandler := ws.NewHandler(hub, messageRepository)
	r.GET("/ws/channel/join/:server_id/:channel_id/:user_id", wsHandler.JoinChannel)

	return r
}
