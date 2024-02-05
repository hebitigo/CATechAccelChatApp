package handler

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	ServerID string `json:"server_id"`
	Name     string `json:"name"`
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
	response := responseRegisterServer{
		ServerID: serverId,
		Name:     request.Name,
	}

	c.JSON(200, response)
}

type requestGetServersByUserID struct {
	UserId string `uri:"user_id" validate:"required"`
}

type responseGetServersByUserID struct {
	ServerID string `json:"server_id"`
	Name     string `json:"name"`
}

func (handler *ServerHandler) GetServersByUserID(c *gin.Context) {
	//https://gin-gonic.com/docs/examples/bind-uri/
	//path pramのvalidate
	log.Printf("path param -> %+v", c.Param("user_Id"))
	var request requestGetServersByUserID
	err := c.BindUri(&request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("failed to bind path param. request -> %+v", requestGetServersByUserID{}))
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
	getServersByUserIDInputDTO := usecase.GetServersByUserIDInputDTO{
		UserId: request.UserId,
	}
	servers, err := handler.usecase.GetServersByUserID(c.Request.Context(), getServersByUserIDInputDTO)
	if err != nil {
		log.Printf("failed to get servers by user_id: %+v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var response []responseGetServersByUserID
	for _, server := range servers {
		response = append(response, responseGetServersByUserID{
			ServerID: server.Id.String(),
			Name:     server.Name,
		})
	}
	c.JSON(200, response)
}

type UserHandler struct {
	usecase usecase.UserUsecaseInterface
}

func NewUserHandler(usecase usecase.UserUsecaseInterface) *UserHandler {
	return &UserHandler{usecase: usecase}
}

type requestUpsertUser struct {
	Id   string `json:"user_id" validate:"required"`
	Name string `json:"name" validate:"required"`
	//https://github.com/go-playground/validator/issues/142#issuecomment-127451987
	Active       *bool  `json:"active" validate:"required"`
	IconImageURL string `json:"icon_image_url"`
}

func (handler *UserHandler) UpsertUser(c *gin.Context) {
	var request requestUpsertUser
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
	upsertUserInputDTO := usecase.UpsertUserInputDTO{
		Id:           request.Id,
		Name:         request.Name,
		Active:       *request.Active,
		IconImageURL: request.IconImageURL,
	}
	err = handler.usecase.UpsertUser(c.Request.Context(), upsertUserInputDTO)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "user upserted successfully"})
}

type ChannelHandler struct {
	usecase usecase.ChannelUsecaseInterface
}

func NewChannelHandler(usecase usecase.ChannelUsecaseInterface) *ChannelHandler {
	return &ChannelHandler{usecase: usecase}
}

type requestRegisterChannel struct {
	ServerId string `json:"server_id" validate:"required,uuid"`
	Name     string `json:"name" validate:"required"`
}

type responseRegisterChannel struct {
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
}

func (handler *ChannelHandler) RegisterChannel(c *gin.Context) {
	var request requestRegisterChannel
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
	//uuidに変換する処理
	serverId, err := uuid.Parse(request.ServerId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	registerChannelInputDTO := usecase.RegisterChannelInputDTO{
		ServerId:    serverId,
		ChannelName: request.Name,
	}
	channelId, err := handler.usecase.RegisterChannel(c.Request.Context(), registerChannelInputDTO)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	response := responseRegisterChannel{
		ChannelID: channelId,
		Name:      request.Name,
	}
	c.JSON(200, response)
}

type requestGetChannelsByServerID struct {
	ServerId string `uri:"server_id" validate:"required,uuid"`
}

type responseGetChannelsByServerID struct {
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
}

func (handler *ChannelHandler) GetChannelsByServerID(c *gin.Context) {
	var request requestGetChannelsByServerID
	err := c.BindUri(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Printf("request -> %+v", request)
	validator := validate.GetValidater()
	err = validator.Struct(request)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("request validation failed:%s", err.Error())})
		return
	}
	serverId, err := uuid.Parse(request.ServerId)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	getChannelsByServerIDInputDTO := usecase.GetChannelsByServerIDInputDTO{
		ServerId: serverId,
	}
	channels, err := handler.usecase.GetChannelsByServerID(c.Request.Context(), getChannelsByServerIDInputDTO)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var response []responseGetChannelsByServerID
	for _, channel := range channels {
		response = append(response, responseGetChannelsByServerID{
			ChannelID: channel.Id.String(),
			Name:      channel.Name,
		})
	}
	c.JSON(200, response)
}
