package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/hebitigo/CATechAccelChatApp/entity"
)

type key string

const (
	txKey key = "tx"
)

func GetInsertQuery(ctx context.Context, db *bun.DB) *bun.InsertQuery {
	if tx, ok := ctx.Value(txKey).(*bun.Tx); ok {
		return tx.NewInsert()
	}
	return db.NewInsert()
}

type BotEndpointRespositoryInterface interface {
	Insert(e entity.BotEndpoint) error
}

// TODO:メソッドの引数でcontext.Contextを受け取るようにする
type BotEndpointRepository struct {
	db  *bun.DB
	ctx context.Context
}

func NewBotEndpointRepository(db *bun.DB, ctx context.Context) *BotEndpointRepository {
	return &BotEndpointRepository{db: db, ctx: ctx}
}

func (repo *BotEndpointRepository) Insert(botEndpoint entity.BotEndpoint) error {
	_, err := repo.db.NewInsert().Model(&botEndpoint).Exec(repo.ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to insert botEndpoint. botEndpoint -> %+v:", botEndpoint))
	}
	return nil
}

type ServerRepositoryInterface interface {
	Insert(ctx context.Context, e entity.Server) (serverId uuid.UUID, err error)
	GetServersByUserID(ctx context.Context, userId string) ([]entity.Server, error)
	GetServer(ctx context.Context, serverId string) (entity.Server, error)
}

type ServerRepository struct {
	db *bun.DB
}

func NewServerRepository(db *bun.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (repo *ServerRepository) Insert(ctx context.Context, e entity.Server) (serverId uuid.UUID, err error) {
	Insert := GetInsertQuery(ctx, repo.db)

	_, err = Insert.Model(&e).Exec(ctx)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, fmt.Sprintf("failed to insert server. server -> %+v:", e))
	}
	return *e.Id, nil
}

func (repo *ServerRepository) GetServersByUserID(ctx context.Context, userId string) ([]entity.Server, error) {
	var servers []entity.Server
	var userServers []entity.UserServer
	err := repo.db.NewSelect().Model(&userServers).Where("user_id = ?", userId).Scan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get userServer by user_id. user_id -> %s", userId))
	}
	if len(userServers) == 0 {
		return nil, nil
	}

	serverIds := make([]string, len(userServers))
	for i, userServer := range userServers {
		serverIds[i] = userServer.ServerId.String()
	}
	//https://bun.uptrace.dev/guide/query-where.html#where-in
	err = repo.db.NewSelect().Model(&servers).Where("id IN (?)", bun.In(serverIds)).Scan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get servers by user_id. user_id -> %s", userId))
	}
	return servers, nil
}

func (repo *ServerRepository) GetServer(ctx context.Context, serverId string) (entity.Server, error) {
	var server entity.Server
	err := repo.db.NewSelect().Model(&server).Where("id = ?", serverId).Scan(ctx)
	if err != nil {
		return entity.Server{}, errors.Wrap(err, fmt.Sprintf("failed to get server by id. server_id -> %s", serverId))
	}
	return server, nil
}

type UserServerRepositoryInterface interface {
	Insert(ctx context.Context, e entity.UserServer) error
}

type UserServerRepository struct {
	db *bun.DB
}

func NewUserServerRepository(db *bun.DB) *UserServerRepository {
	return &UserServerRepository{db: db}
}

func (repo *UserServerRepository) Insert(ctx context.Context, e entity.UserServer) error {
	Insert := GetInsertQuery(ctx, repo.db)

	_, err := Insert.Model(&e).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to insert userServer. userserver -> %+v:", e))
	}
	return nil
}

type ChannelRepositoryInterface interface {
	Insert(ctx context.Context, e entity.Channel) (channelId uuid.UUID, err error)
	GetChannelsByServerID(ctx context.Context, serverId uuid.UUID) ([]entity.Channel, error)
}

type ChannelRepository struct {
	db *bun.DB
}

func NewChannelRepository(db *bun.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (repo *ChannelRepository) Insert(ctx context.Context, e entity.Channel) (channelId uuid.UUID, err error) {
	Insert := GetInsertQuery(ctx, repo.db)

	_, err = Insert.Model(&e).Exec(ctx)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, fmt.Sprintf("failed to insert channel. channel -> %+v:", e))
	}
	return *e.Id, nil
}

func (repo *ChannelRepository) GetChannelsByServerID(ctx context.Context, serverId uuid.UUID) ([]entity.Channel, error) {
	var channels []entity.Channel
	err := repo.db.NewSelect().Model(&channels).Where("server_id = ?", serverId).Scan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get channels by server_id. server_id -> %s", serverId))
	}
	return channels, nil
}

type TxRepositoryInterface interface {
	DoInTx(ctx context.Context, f func(ctx context.Context) error) error
}

type TxRepository struct {
	db *bun.DB
}

func NewTxRepository(db *bun.DB) *TxRepository {
	return &TxRepository{db: db}
}

func (repos TxRepository) DoInTx(ctx context.Context, f func(ctx context.Context) error) error {
	var done bool

	tx, err := repos.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to begin tx. tx -> %v:", err))
	}

	ctx = context.WithValue(ctx, txKey, &tx)

	t, ok := ctx.Value(txKey).(*bun.Tx)
	log.Println("tx: ", t, "ok: ", ok)

	defer func() {
		if !done {
			err = tx.Rollback()
			if err != nil {
				log.Printf("failed to rollback: %v", err)
			}
		}
	}()

	err = f(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to execute function in tx:")
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit tx:")
	}
	done = true
	return nil
}

type UserRepositoryInterface interface {
	Upsert(ctx context.Context, e entity.User) error
	GetUser(ctx context.Context, userId string) (entity.User, error)
	CheckUserExist(ctx context.Context, userId string) error
}

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) GetUser(ctx context.Context, userId string) (entity.User, error) {
	var user entity.User
	err := repo.db.NewSelect().Model(&user).Where("id = ?", userId).Scan(ctx)
	if err != nil {
		return entity.User{}, errors.Wrap(err, fmt.Sprintf("failed to get user by id. user_id -> %s", userId))
	}
	return user, nil
}

func (repo *UserRepository) Upsert(ctx context.Context, e entity.User) error {
	_, err := repo.db.NewInsert().Model(&e).On("CONFLICT (id) DO UPDATE").Set("name = EXCLUDED.name").Set("active = EXCLUDED.active").Set("icon_image_url = EXCLUDED.icon_image_url").Exec(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to insert user. user -> %+v:", e))
	}
	return nil
}

func (repo *UserRepository) CheckUserExist(ctx context.Context, userId string) error {
	var user entity.User
	err := repo.db.NewSelect().Model(&user).Where("id = ?", userId).Scan(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get user by id. user_id -> %s", userId))
	}
	return nil
}

type MessageRepositoryInterface interface {
	Insert(ctx context.Context, e entity.Message) (time.Time, uuid.UUID, error)
	GetMessagesWithUser(ctx context.Context, channelId uuid.UUID) ([]entity.MessageWithUser, error)
}

type MessageRepository struct {
	db *bun.DB
}

func NewMessageRepository(db *bun.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (repo *MessageRepository) Insert(ctx context.Context, e entity.Message) (time.Time, uuid.UUID, error) {
	_, err := repo.db.NewInsert().Model(&e).Exec(ctx)
	if err != nil {
		return time.Time{}, uuid.UUID{}, errors.Wrap(err, fmt.Sprintf("failed to insert message. message -> %+v:", e))
	}
	return e.CreatedAt, *e.Id, nil
}

func (repo *MessageRepository) GetMessagesWithUser(ctx context.Context, channelId uuid.UUID) ([]entity.MessageWithUser, error) {
	var messages []entity.MessageWithUser
	// err := repo.db.NewSelect().Table("messages AS message").ColumnExpr("*").ColumnExpr("name as user_name,user.icon_image_url as user_icon_image_url").Join("JOIN users as user ON message.user_id = user.id").Where("message.channel_id = ?", channelId).Scan(ctx, &messages)
	err := repo.db.NewSelect().TableExpr("messages AS message").ColumnExpr("message.*").ColumnExpr("u.name as user_name,u.icon_image_url as user_icon_image_url").Join("INNER JOIN users AS u ON message.user_id = u.id").Where("message.channel_id = ?", channelId).Scan(ctx, &messages)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to get messages by channel_id. channel_id -> %s", channelId))
	}
	return messages, nil
}
