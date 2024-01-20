package handler

import (
	"CATechAccelChatApp/entity"
	"CATechAccelChatApp/usecase"

	"github.com/gin-gonic/gin"
)

type botEndpointHandler struct {
	usecase usecase.BotEndpointUsecaseInterface
}

func NewBotEndpointHandler(usecase usecase.BotEndpointUsecaseInterface) *botEndpointHandler {
	return &botEndpointHandler{usecase: usecase}
}

//	### POST /registerBotEndpoint
//
// bot のエンドポイントを登録
//
// ヘッダ
//
// ```
// Content-Type: application/json
// Authorization: Bearer {jwt}
// ```
//
// ボディ
//
// ```
//
//	{
//	   "id": "string",
//	   "name": "string",
//	   "icon_url": "string",
//	   "endpoint": "string",
//	}
//
// ```
func (handler *botEndpointHandler) RegisterBotEndpoint(ctx *gin.Context) {
	var botEndpoint entity.BotEndpoint
	if err := ctx.BindJSON(&botEndpoint); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := handler.usecase.RegisterBotEndpoint(&botEndpoint); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"message": "bot endpoint registered successfully"})
}
