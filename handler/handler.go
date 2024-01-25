package handler

import (
	"fmt"

	"github.com/hebitigo/CATechAccelChatApp/usecase"
	validate "github.com/hebitigo/CATechAccelChatApp/util"

	"github.com/hebitigo/CATechAccelChatApp/entity"

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
	validator := validate.GetValidater()
	if err := validator.Struct(botEndpoint); err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("botEndpointParams validation failed:%s", err.Error())})
		return
	}

	if err := handler.usecase.RegisterBotEndpoint(&botEndpoint); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"message": "bot endpoint registered successfully"})
}

type ServerHandler struct {
	usecase usecase.ServerUsecaseInterface
}

func NewServerHandler(usecase usecase.ServerUsecaseInterface) *ServerHandler {
	return &ServerHandler{usecase: usecase}
}

type requestRegisterServer struct {
	UserId string `json:"user_id" validate:"required"`
	Name   string `json:"name" validate:"required"`
}

type responseRegisterServer struct {
	ServerId string `json:"server_id"`
}

func (handler *ServerHandler) RegisterServer(c *gin.Context) {
	var request requestRegisterServer
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	if err := validator.Struct(request); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("request validation failed:%s", err.Error())})
		return
	}
	registerServerDto := usecase.RegisterServerInputDTO{
		ServerName: request.Name,
		UserId:     request.UserId,
	}
	serverId, err := handler.usecase.RegisterServer(c.Request.Context(), registerServerDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, responseRegisterServer{ServerId: serverId})
}

type UserHandler struct {
	usecase usecase.UserUsecaseInterface
}

func NewUserHandler(usecase usecase.UserUsecaseInterface) *UserHandler {
	return &UserHandler{usecase: usecase}
}

type requestRegisterUser struct {
	Id           string `json:"user_id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Active       bool   `json:"active" validate:"required"`
	IconImageURL string `json:"icon_image_url" validate:"required"`
}

func (handler *UserHandler) RegisterUser(c *gin.Context) {
	var request requestRegisterUser
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	if err := validator.Struct(request); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("request validation failed:%s", err.Error())})
		return
	}
	registerUserInputDto := usecase.RegisterUserInputDTO{
		Id:           request.Id,
		Name:         request.Name,
		Active:       request.Active,
		IconImageURL: request.IconImageURL,
	}
	if err := handler.usecase.RegisterUser(c.Request.Context(), registerUserInputDto); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "user registered successfully"})
}
