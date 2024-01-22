package entity

import "time"

// foreing key制約
// https://bun.uptrace.dev/guide/query-create-table.html#api

type BotEndpoint struct {
	ID       string `json:"id" bun:"id,pk,uuid,default:gen_random_uuid()" validate:"isempty"`
	ENDPOINT string `json:"endpoint" bun:"endpoint" validate:"required"`
	NAME     string `json:"name" bun:"name" validate:"required"`
	ICON_URL string `json:"icon_url" bun:"icon_url" validate:"required"`
}

type Server struct {
	ID   string `json:"server_id" bun:"id,pk,uuid,default:gen_random_uuid()"`
	NAME string `json:"name" bun:"name"`
}

type Channel struct {
	ID       string `json:"channel_id" bun:"id,pk,uuid,default:gen_random_uuid()"`
	ServerID string `json:"server_id" bun:"server_id"`
	NAME     string `json:"name" bun:"name"`
}

type User struct {
	ID           string `json:"user_id" bun:"id,pk"`
	Name         string `json:"name" bun:"name"`
	Active       bool   `json:"active" bun:"active"`
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
	ID            string    `json:"message_id" bun:"id,pk,uuid,default:gen_random_uuid()"`
	ChannelID     string    `json:"channel_id" bun:"channel_id,notnull"`
	UserID        string    `json:"user_id" bun:"user_id"`
	BotEndpointID string    `json:"bot_endpoint_id" bun:"bot_endpoint_id"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at"`
	IsBot         bool      `json:"is_bot" bun:"is_bot,notnull"`
}
