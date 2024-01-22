package repository

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"github.com/hebitigo/CATechAccelChatApp/entity"
)

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
	fmt.Println("inserting information:", botEndpoint)
	_, err := repo.db.NewInsert().Model(botEndpoint).Exec(repo.ctx)
	fmt.Println("inserted bot endpoint id: ", botEndpoint.ID)
	if err != nil {
		return err
	}
	return nil
}
