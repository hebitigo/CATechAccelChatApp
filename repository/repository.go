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
	Insert(ctx context.Context, e entity.Server) (serverId *uuid.UUID, err error)
}

type ServerRepository struct {
	db *bun.DB
}

func NewServerRepository(db *bun.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (repo *ServerRepository) Insert(ctx context.Context, e entity.Server) (serverId *uuid.UUID, err error) {
	Insert := GetInsertQuery(ctx, repo.db)

	_, err = Insert.Model(&e).Exec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to insert server. server -> %+v:", e))
	}
	return e.Id, nil
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
	Insert(ctx context.Context, e entity.Channel) error
}

type ChannelRepository struct {
	db *bun.DB
}

func NewChannelRepository(db *bun.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (repo *ChannelRepository) Insert(ctx context.Context, e entity.Channel) error {
	Insert := GetInsertQuery(ctx, repo.db)

	_, err := Insert.Model(&e).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to insert channel. channel -> %+v:", e))
	}
	return nil
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
	Insert(ctx context.Context, e entity.User) error
}

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Insert(ctx context.Context, e entity.User) error {
	_, err := repo.db.NewInsert().Model(&e).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to insert user. user -> %+v:", e))
	}
	return nil
}

type MessageRepositoryInterface interface {
	Insert(ctx context.Context, e entity.Message) (time.Time, error)
}

type MessageRepository struct {
	db *bun.DB
}

func NewMessageRepository(db *bun.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (repo *MessageRepository) Insert(ctx context.Context, e entity.Message) (time.Time, error) {
	_, err := repo.db.NewInsert().Model(&e).Exec(ctx)
	if err != nil {
		return time.Time{}, errors.Wrap(err, fmt.Sprintf("failed to insert message. message -> %+v:", e))
	}
	return e.CreatedAt, nil
}
