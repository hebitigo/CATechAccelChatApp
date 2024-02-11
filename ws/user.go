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
	"github.com/hebitigo/CATechAccelChatApp/util"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type User struct {
	UserID      string
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	messageRepo repository.MessageRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	ctx         context.Context
}

type actionType string

const (
	chatMessageAction  actionType = "chat_message"
	addChannelAction   actionType = "add_channel"
	userActivateAction actionType = "user_activate"
	errorAction        actionType = "error"
)

type incomingChatMessageInfo struct {
	ServerId  string `json:"server_id" validate:"required,uuid"`
	ChannelId string `json:"channel_id" validate:"required,uuid"`
	Message   string `json:"message" validate:"required"`
}

type outgoingChatMessageInfo struct {
	UserName         string    `json:"user_name"`
	UserIconImageURL string    `json:"user_icon_image_url"`
	ServerId         string    `json:"server_id"`
	ChannelId        string    `json:"channel_id"`
	Message          string    `json:"message"`
	CreatedAt        time.Time `json:"created_at"`
}

type channelInfo struct {
	Name      string `json:"name"`
	ServerId  string `json:"server_id"`
	ChannelId string `json:"channel_id"`
}

type returnError struct {
	Message string `json:"message"`
}

type readMessage struct {
	ActionType actionType      `json:"action_type" validate:"required"`
	Payload    json.RawMessage `json:"payload" validate:"required"`
}

type Payload interface {
	outgoingChatMessageInfo | incomingChatMessageInfo | channelInfo | returnError
}

type SendMessage struct {
	ActionType actionType  `json:"action_type"`
	Payload    interface{} `json:"payload"`
}

func returnSendMessage[P Payload](at actionType, p P) SendMessage {
	return SendMessage{
		ActionType: at,
		Payload:    p,
	}
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
	validator := util.GetValidater()
Loop:
	for {
		_, byteMessage, err := u.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				err = errors.Wrap(err, fmt.Sprintf("read message from websocket is failed.err is unexpected error. byteMessage -> %+v", byteMessage))
				log.Printf("%+v", err)
			}
			break
		}

		var readMessage readMessage
		err = json.Unmarshal(byteMessage, &readMessage)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("cant unmarshal byteMessage from Websocket. byteMessage -> %+v", byteMessage))
			log.Printf("%+v", err)

			sendWebsocketError(u.conn, err)
			break
		}
		err = validator.Struct(readMessage)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("readMessage is invalid. readMessage -> %+v", readMessage))
			log.Printf("%+v", err)
			sendWebsocketError(u.conn, err)
			break
		}

		switch readMessage.ActionType {
		case chatMessageAction:
			var chatMessageInfo incomingChatMessageInfo
			err := json.Unmarshal(readMessage.Payload, &chatMessageInfo)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("cant unmarshal chatMessageInfo from readMessage.Payload. readMessage.Payload -> %+v", readMessage.Payload))
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				break
			}
			err = validator.Struct(chatMessageInfo)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("chatMessageInfo is invalid. chatMessageInfo -> %+v", chatMessageInfo))
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				break
			}
			channelId, err := uuid.Parse(chatMessageInfo.ChannelId)
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("cant parse channelId. channelId -> %s", chatMessageInfo.ChannelId))
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				break
			}

			user, err := u.userRepo.GetUser(u.ctx, u.UserID)
			if err != nil {
				log.Printf("failed to get user info: %+v", err)
				sendWebsocketError(u.conn, err)
				break Loop

			}

			message := entity.Message{
				UserId:        user.Id,
				ChannelId:     channelId,
				IsBot:         false,
				Message:       chatMessageInfo.Message,
				BotEndpointId: nil,
			}
			createdAt, err := u.messageRepo.Insert(u.ctx, message)
			if err != nil {
				log.Printf("failed to insert message provided by websocket: %+v", err)
				sendWebsocketError(u.conn, err)
				break

			}

			returnChatMessageInfo := outgoingChatMessageInfo{
				UserName:         user.Name,
				UserIconImageURL: user.IconImageURL,
				CreatedAt:        createdAt,
				ServerId:         chatMessageInfo.ServerId,
				ChannelId:        chatMessageInfo.ChannelId,
				Message:          chatMessageInfo.Message,
			}
			bytes, err := json.Marshal(returnSendMessage[outgoingChatMessageInfo](
				chatMessageAction,
				returnChatMessageInfo,
			))
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("cant marshal returnChatMessageInfo. returnChatMessageInfo -> %+v", returnChatMessageInfo))
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				break
			}
			u.hub.broadcast <- bytes
		default:
			err = errors.New(fmt.Sprintf("unexpected actionType. actionType -> %s", readMessage.ActionType))
			log.Printf("%+v", err)
			sendWebsocketError(u.conn, err)
			break Loop
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
		case payload, ok := <-u.send:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				u.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := u.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				err = errors.Wrap(err, "failed to get next writer:")
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				return
			}
			w.Write(payload)

			n := len(u.send)
			for i := 0; i < n; i++ {
				payload := <-u.send
				w.Write(payload)
			}
			err = w.Close()
			if err != nil {
				err = errors.Wrap(err, "failed to close writer:")
				log.Printf("%+v", err)
				sendWebsocketError(u.conn, err)
				return
			}
		case <-ticker.C:
			u.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Printf("ping and deadline is set")
			err := u.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

func sendWebsocketError(conn *websocket.Conn, err error) {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	bytes, err := json.Marshal(returnSendMessage[returnError](errorAction, returnError{Message: err.Error()}))
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("cant marshal error message. err -> %+v", err))
		log.Printf("%+v", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to write message to websocket. message -> %+v", bytes))
		log.Printf("%+v", err)
		return
	}
}
