package ws

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
		c.JSON(500, gin.H{"message": fmt.Sprintf("failed to upgrade http to websocket: %v", err)})
		return
	}

	//check uuid
	channelId, err := uuid.Parse(c.Param("channel_Id"))
	if err != nil {
		c.JSON(400, gin.H{"message": fmt.Sprintf("channel_Id is not uuid: %v", err)})
		return
	}
	serverId, err := uuid.Parse(c.Param("server_Id"))
	if err != nil {
		c.JSON(400, gin.H{"message": fmt.Sprintf("server_Id is not uuid: %v", err)})
		return
	}
	user := &User{
		ChannelID:   channelId,
		ServerID:    serverId,
		UserID:      c.Param("user_Id"),
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
