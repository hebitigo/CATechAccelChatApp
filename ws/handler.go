package ws

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/hebitigo/CATechAccelChatApp/repository"
)

type Handler struct {
	hub         *Hub
	messageRepo repository.MessageRepositoryInterface
	userRepo repository.UserRepositoryInterface
}

func NewHandler(hub *Hub, messageRepo repository.MessageRepositoryInterface,userRepo repository.UserRepositoryInterface) *Handler {
	return &Handler{hub: hub, messageRepo: messageRepo,userRepo:userRepo}
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

	user := &User{
		UserID:      c.Param("user_id"),
		hub:         handler.hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		ctx:         context.Background(),
		messageRepo: handler.messageRepo,
		userRepo: handler.userRepo,
	}

	handler.hub.register <- user
	go user.writePump()
	go user.readPump()

}
