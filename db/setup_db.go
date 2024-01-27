package db

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
	// bunは構造体からテーブルを作成する際はデフォルトで構造体の名前の複数形のテーブルを作成する
	//https://bun.uptrace.dev/guide/golang-orm.html#defining-models　で構造体にbun.BaseModelを埋め込んでタグを指定することで
	//テーブル名を変更できるが、entityの構造体に外部ライブラリを埋め込みたくなかったので、テーブル名はデフォルトのままにしている
	_, err := db.NewCreateTable().Model((*entity.BotEndpoint)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create bot_endpoint table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.Server)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create server table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.Channel)(nil)).IfNotExists().ForeignKey("(server_id) REFERENCES servers (id) ON DELETE CASCADE").Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create channel table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.User)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create user table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.Message)(nil)).IfNotExists().ForeignKey("(user_id) REFERENCES users (id) ON DELETE CASCADE").ForeignKey("(bot_endpoint_id) REFERENCES bot_endpoints (id) ON DELETE CASCADE").ForeignKey("(channel_id) REFERENCES channels (id) ON DELETE CASCADE").Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create message table: %v", err)
	}

	//メッセージテーブルに制約があるかどうかを"bot_id_or_user_id"という名前の制約が掛かっているカラムの数を取得して確認する
	MessagesTableConstraintCount, err := db.NewSelect().Table("information_schema.constraint_column_usage").Where("table_name =? AND constraint_name = ?", "messages", "bot_id_or_user_id").Count(ctx)
	if err != nil {
		log.Fatalf("failed to check if messages table constraint exists: %v", err)
	}
	existMessagesTableConstraint := MessagesTableConstraintCount > 0
	if !existMessagesTableConstraint {
		_, err = db.Exec(`
		ALTER TABLE messages
		ADD CONSTRAINT bot_id_or_user_id
		CHECK ((is_bot = true AND bot_endpoint_id IS NOT NULL) OR (is_bot = false AND user_id IS NOT NULL));`)
		if err != nil {
			log.Fatalf("failed to add constraint to message table: %v", err)
		}
	}
	_, err = db.NewCreateTable().Model((*entity.ServerBotEndpoint)(nil)).IfNotExists().ForeignKey("(server_id) REFERENCES servers (id) ON DELETE CASCADE").ForeignKey("(bot_endpoint_id) REFERENCES bot_endpoints (id) ON DELETE CASCADE").Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create server_bot_endpoint table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.UserServer)(nil)).IfNotExists().ForeignKey("(user_id) REFERENCES users (id) ON DELETE CASCADE").ForeignKey("(server_id) REFERENCES servers (id) ON DELETE CASCADE").Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create user_server table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.ReactionType)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create reaction_type table: %v", err)
	}
	_, err = db.NewCreateTable().Model((*entity.UserReaction)(nil)).IfNotExists().ForeignKey("(user_id) REFERENCES users (id) ON DELETE CASCADE").ForeignKey("(message_id) REFERENCES messages (id) ON DELETE CASCADE").ForeignKey("(reaction_type_id) REFERENCES reaction_types (id) ON DELETE CASCADE").Exec(ctx)
	if err != nil {
		log.Fatalf("failed to create user_reaction table: %v", err)
	}
}
