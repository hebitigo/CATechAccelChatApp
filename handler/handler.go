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

func (handler *botEndpointHandler) RegisterBotEndpoint(c *gin.Context) {
	var request requestRegisterBotEndpoint
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	validator := validate.GetValidater()
	err = validator.Struct(request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("requestRegisterBotEndpoint validation failed. botEndpoint -> %+v", request))
		log.Printf("%+v", err)
		c.JSON(400, gin.H{"error": err.Error()})
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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "bot endpoint registered successfully"})
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
	c.JSON(200, gin.H{"message": "server registered successfully", "server_id": serverId})
}

type requestGetServersByUserID struct {
	UserId string `uri:"user_Id" validate:"required"`
}

func (handler *ServerHandler) GetServersByUserID(c *gin.Context) {
	//https://gin-gonic.com/docs/examples/bind-uri/
	//path pramのvalidate
	log.Printf("path param -> %+v", c.Param("user_Id"))
	var request requestGetServersByUserID
	err := c.BindUri(&request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("failed to bind path param. request -> %+v", request))
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
	c.JSON(200, servers)
}

type requestCreateInvitationByJWT struct {
	UserId   string `json:"user_id" validate:"required"`
	ServerId string `json:"server_id" validate:"required,uuid"`
}

type responseCreateInvitationByJWT struct {
	Token []byte `json:"token"` //jwt
}

func (handler *ServerHandler) CreateInvitationByJWT(c *gin.Context) {
	//https://gin-gonic.com/docs/examples/bind-uri/
	//path pramのvalidate
	var request requestCreateInvitationByJWT
	err := c.BindJSON(&request)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("failed to bind path param. request -> %+v", request))
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
	createInvitationByJWTDTO := usecase.CreateInvitationByJWTInputDTO{
		UserId:   request.UserId,
		ServerId: request.ServerId,
	}
	signed, err := handler.usecase.CreateInvitationByJWT(createInvitationByJWTDTO)
	if err != nil {
		log.Printf("failed to get servers by user_id: %+v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	response := responseCreateInvitationByJWT{
		Token: signed,
	}
	c.JSON(200, response)
}

type requestJoinServerByInvitation struct {
	Token  []byte `json:"token" validate:"required"` //jwt
	UserId string `json:"user_id" validate:"required"`
}

type responseJoinServerByInvitation struct {
	ServerID string `json:"server_id"`
	Name     string `json:"name"`
}

func (handler *ServerHandler) JoinServerByInvitation(c *gin.Context) {
	var request requestJoinServerByInvitation
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
	authAndAddUserInputDTO := usecase.AuthAndAddUserInputDTO{
		Token:  request.Token,
		UserId: request.UserId,
	}
	server, err := handler.usecase.AuthAndAddUser(c.Request.Context(), authAndAddUserInputDTO)
	if err != nil {
		log.Printf("failed to auth and add user: %+v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	response := responseJoinServerByInvitation{
		ServerID: server.Id.String(),
		Name:     server.Name,
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
