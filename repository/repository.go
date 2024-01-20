package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/hebitigo/CATechAccelChatApp/entity"
)

func GetDBConnection(ctx context.Context) *bun.DB {

	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	sqldb, err := sql.Open("pg", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	DBInit(db, ctx)

	return db

}

func DBInit(db *bun.DB, ctx context.Context) {
	//	テーブルがない場合は作成する
	_, err := db.NewCreateTable().Model((*entity.BotEndpoint)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create bot_endpoint table: %v", err)
	}

}

type BotEndpointRespositoryInterface interface {
	Insert(e *entity.BotEndpoint) error
}

type BotEndpointRepository struct {
	db  *bun.DB
	ctx context.Context
}

func NewBotEndpointRepository(db *bun.DB, ctx context.Context) *BotEndpointRepository {
	return &BotEndpointRepository{db: db, ctx: ctx}
}

func (repo *BotEndpointRepository) Insert(botEndpoint *entity.BotEndpoint) error {
	_, err := repo.db.NewInsert().Model(botEndpoint).Exec(repo.ctx)
	if err != nil {
		return err
	}
	return nil
}
