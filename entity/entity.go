package entity

import (
	"time"

	"github.com/google/uuid"
)

// foreing key制約
// https://bun.uptrace.dev/guide/query-create-table.html#api

type BotEndpoint struct {
	Id       string `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" validate:"isempty"`
	Endpoint string `json:"endpoint" bun:"endpoint,notnull" validate:"required"`
	Name     string `json:"name" bun:"name,notnull" validate:"required"`
	IconURL  string `json:"icon_url" bun:"icon_url,notnull" validate:"required"`
}

type Server struct {
	Id   *uuid.UUID `json:"server_id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name string     `json:"name" bun:"name,notnull"`
}

type Channel struct {
	Id       *uuid.UUID `json:"channel_id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ServerId *uuid.UUID `json:"server_id" bun:"server_id,unique:serverIdAndChannelName,notnull,type:uuid"` //FK
	Name     string     `json:"name" bun:"name,unique:serverIdAndChannelName,notnull"`
}

type User struct {
	Id           string `json:"user_id" bun:"id,pk"`
	Name         string `json:"name" bun:"name,notnull"`
	Active       bool   `json:"active" bun:"active,notnull"`
	IconImageURL string `json:"icon_image_url" bun:"icon_image_url"`
}

// IsBotによってbotからのメッセージかどうかを判定する
// IsBotがtrueの場合はBotEndpointIDが必須, falseの場合はUserIDが必須

// Message構造体のテーブルを作成した後にSQLで制約を追加する
// _, err = db.Exec(`
// ALTER TABLE messages
// ADD CONSTRAINT bot_id_or_user_id
// CHECK ((is_bot = true AND bot_endpoint_id IS NOT NULL) OR (is_bot = false AND user_id IS NOT NULL));`)
type Message struct {
	Id            *uuid.UUID `json:"message_id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ChannelId     *uuid.UUID `json:"channel_id" bun:"channel_id,notnull,type:uuid"`   //FK
	UserId        string     `json:"user_id" bun:"user_id"`                           //FK
	BotEndpointId *uuid.UUID `json:"bot_endpoint_id" bun:"bot_endpoint_id,type:uuid"` //FK
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	Message       string     `json:"message" bun:"message,notnull"`
	IsBot         bool       `json:"is_bot" bun:"is_bot,notnull"`
}

type ServerBotEndpoint struct {
	ServerId      string `json:"server_id" bun:"server_id,pk,type:uuid"`             //FK
	BotEndpointId string `json:"bot_endpoint_id" bun:"bot_endpoint_id,pk,type:uuid"` //FK
}

type UserServer struct {
	UserId   string     `json:"user_id" bun:"user_id,pk"`               //FK
	ServerId *uuid.UUID `json:"server_id" bun:"server_id,pk,type:uuid"` //FK
}

type UserReaction struct {
	Id             string `json:"user_reaction_id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	MessageId      string `json:"message_id" bun:"message_id,notnull,type:uuid"`             //FK
	UserId         string `json:"user_id" bun:"user_id,notnull"`                             //FK
	ReactionTypeId string `json:"reaction_type_id" bun:"reaction_type_id,notnull,type:uuid"` //FK
}

type ReactionType struct {
	Id    string `json:"reaction_type_id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Emoji string `json:"emoji" bun:"emoji,notnull"`
}
