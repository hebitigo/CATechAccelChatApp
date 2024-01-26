package handler

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hebitigo/CATechAccelChatApp/usecase"
	validate "github.com/hebitigo/CATechAccelChatApp/util"
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

type requestRegisterBotEndpoint struct {
	Name     string `json:"name" validate:"required"`
	IconURL  string `json:"icon_url" validate:"required"`
	Endpoint string `json:"endpoint" validate:"required"`
}

func (handler *botEndpointHandler) RegisterBotEndpoint(ctx *gin.Context) {
	var request requestRegisterBotEndpoint
	err := ctx.BindJSON(&request)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	err = validator.Struct(request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("requestRegisterBotEndpoint validation failed. botEndpoint -> %+v", request))
		log.Printf("%+v", err)
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	registerBotEndpointDto := usecase.RegisterBotEndpointInputDTO{
		Name:     request.Name,
		IconURL:  request.IconURL,
		Endpoint: request.Endpoint,
	}

	err = handler.usecase.RegisterBotEndpoint(registerBotEndpointDto)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("failed to register bot endpoint. botEndpoint -> %+v", request))
		log.Printf("%+v", err)
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
	err := c.BindJSON(&request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("failed to bind json. request -> %+v", request))
		log.Printf("%+v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	err = validator.Struct(request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("request validation failed. request -> %+v", request))
		log.Printf("%+v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	registerServerDto := usecase.RegisterServerInputDTO{
		ServerName: request.Name,
		UserId:     request.UserId,
	}
	serverId, err := handler.usecase.RegisterServer(c.Request.Context(), registerServerDto)
	if err != nil {
		log.Printf("failed to register server: %+v", err)
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
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	err = validator.Struct(request)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("request validation failed:%s", err.Error())})
		return
	}
	registerUserInputDto := usecase.RegisterUserInputDTO{
		Id:           request.Id,
		Name:         request.Name,
		Active:       request.Active,
		IconImageURL: request.IconImageURL,
	}
	err = handler.usecase.RegisterUser(c.Request.Context(), registerUserInputDto)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "user registered successfully"})
}
