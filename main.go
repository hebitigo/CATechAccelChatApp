package main

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func main() {
	//	### POST /registerBotEndpoint
	//
	//bot のエンドポイントを登録
	//
	//ヘッダ
	//
	//```
	//Content-Type: application/json
	//Authorization: Bearer {jwt}
	//```
	//
	//ボディ
	//
	//```
	//{
	//    "id": "string",
	//    "name": "string",
	//    "icon_url": "string",
	//    "endpoint": "string",
	//}
	//```
	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	sqldb, err := sql.Open("pg", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer sqldb.Close()
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	err = db.Ping()
	ctx := context.Background()
	DB := NewDB(db, ctx)

	DB.Init(ctx)
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	r := gin.Default()
	r.POST("/registerBotEndpoint", DB.registerBotEndpoint)
	r.Run(":8080")
}

type DB struct {
	db  *bun.DB
	ctx context.Context
}

type BOT_ENDPOINT struct {
	ID       string `json:"id" bun:"id,pk,autoincrement"`
	ENDPOINT string `json:"endpoint" bun:"endpoint"`
	NAME     string `json:"name" bun:"name"`
	ICON_URL string `json:"icon_url" bun:"icon_url"`
}

func (db *DB) Init(ctx context.Context) {
	//	テーブルがない場合は作成する
	_, err := db.db.NewCreateTable().Model((*BOT_ENDPOINT)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create bot_endpoint table: %v", err)
	}

}

func NewDB(db *bun.DB, ctx context.Context) *DB {
	return &DB{db: db, ctx: ctx}
}

func (db *DB) registerBotEndpoint(c *gin.Context) {
	//	TODO:clerkのGO SDKを使ってjwtの検証を行う
	var botEndpoint BOT_ENDPOINT
	err := c.BindJSON(&botEndpoint)
	if err != nil {
		log.Printf("failed to bind json in registering BotEndpoint: %v", err)
		c.JSON(500, gin.H{"message": "failed to bind json in registering BotEndpoint"})
		return
	}
	//	登録する
	_, err = db.db.NewInsert().Model(&botEndpoint).Exec(db.ctx)
	if err != nil {
		log.Printf("failed to insert BotEndpoint: %v", err)
		c.JSON(500, gin.H{"message": "failed to insert BotEndpoint"})
		return
	}
	c.JSON(200, gin.H{"message": "bot endpoint registered successfully"})

}
