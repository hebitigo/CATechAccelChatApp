package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/hebitigo/CATechAccelChatApp/entity"
	"github.com/hebitigo/CATechAccelChatApp/repository"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type User struct {
	ChannelID   uuid.UUID
	ServerID    uuid.UUID
	UserID      string
	hub         *Hub
	conn        *websocket.Conn
	send        chan broadcastMessage
	messageRepo repository.MessageRepositoryInterface
	ctx         context.Context
}

type ReturnMessage struct {
	Name         string    `json:"name"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
	IconImageURL string    `json:"icon_image_url"`
}

func (u *User) readPump() {
	defer func() {
		u.hub.unregister <- u
		u.conn.Close()
	}()
	u.conn.SetReadLimit(maxMessageSize)
	u.conn.SetReadDeadline(time.Now().Add(pongWait))

	//https://websockets.spec.whatwg.org/#ping-and-pong-frames
	//WHATWGによってwebで提供されているWebsocket APIの仕様によると
	//WebSocket APIにはpingとpongのAPIが提供されてなくて、
	//>It is assumed that servers will solicit pongs whenever appropriate for the server’s needs:
	//を見る限り、サーバーが必要に応じてpongを要求するということなので、
	//フロント側ではpingとpongの処理を書かなくてもいいということなのかなと思う

	//サーバー側で一定間隔でpingフレームの送信を要求して、
	//一定時間内にpongフレームを受け取れない場合は接続を切断するという処理を書く
	u.conn.SetPongHandler(func(string) error {
		u.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, byteMessage, err := u.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				err = errors.Wrap(err, fmt.Sprintf("read message from websocket is failed.err is unexpected error. byteMessage -> %+v", byteMessage))
				log.Printf("%+v", err)
			}
			break
		}

		stringMessage := string(byteMessage)

		message := entity.Message{
			UserId:        u.UserID,
			ChannelId:     &u.ChannelID,
			IsBot:         false,
			Message:       stringMessage,
			BotEndpointId: nil,
		}

		createdAt, err := u.messageRepo.Insert(u.ctx, message)
		if err != nil {
			log.Printf("failed to insert message provided by websocket: %+v", err)
			break
		}

		returnMessage, err := json.Marshal(ReturnMessage{
			Name:      u.UserID,
			Message:   stringMessage,
			CreatedAt: createdAt,
		})
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("cant marshal Channel Message. returnMessage -> %+v", returnMessage))
			log.Printf("%v", err)
			break
		}
		u.hub.broadcast <- broadcastMessage{
			serverId:    serverId(u.ServerID),
			channelId:   channelId(u.ChannelID),
			jsonMessage: returnMessage,
		}
	}

}

func (u *User) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		u.conn.Close()
	}()
	for {
		select {
		case broadcastMessage, ok := <-u.send:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				u.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := u.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				err = errors.Wrap(err, "failed to get next writer:")
				log.Printf("%+v", err)
				return
			}
			w.Write(broadcastMessage.jsonMessage)

			n := len(u.send)
			for i := 0; i < n; i++ {
				broadcastMessage := <-u.send
				w.Write(broadcastMessage.jsonMessage)
			}
			err = w.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close writer:")
				log.Printf("%+v", err)
				return
			}
		case <-ticker.C:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := u.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}
