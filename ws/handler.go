package ws

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/hebitigo/CATechAccelChatApp/repository"
)

type Handler struct {
	hub         *Hub
	messageRepo repository.MessageRepositoryInterface
}

func NewHandler(hub *Hub, repository repository.MessageRepositoryInterface) *Handler {
	return &Handler{hub: hub, messageRepo: repository}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (handler *Handler) JoinChannel(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err := errors.Wrap(err, "failed to upgrade http to websocket")
		log.Printf("%+v", err)
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	//check uuid
	channelId, err := uuid.Parse(c.Param("channel_id"))
	if err != nil {
		err := errors.Wrap(err, "channel_id is not uuid")
		log.Printf("%+v", err)
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	serverId, err := uuid.Parse(c.Param("server_id"))
	if err != nil {
		err := errors.Wrap(err, "server_id is not uuid")
		log.Printf("%+v", err)
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	user := &User{
		ChannelID:   channelId,
		ServerID:    serverId,
		UserID:      c.Param("user_id"),
		hub:         handler.hub,
		conn:        conn,
		send:        make(chan broadcastMessage, 256),
		ctx:         context.Background(),
		messageRepo: handler.messageRepo,
	}

	handler.hub.register <- user
	go user.writePump()
	go user.readPump()

}
