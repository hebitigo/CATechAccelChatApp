package usecase

import (
	"CATechAccelChatApp/entity"
	"CATechAccelChatApp/repository"
)

type BotEndpointUsecaseInterface interface {
	RegisterBotEndpoint(botEndpoint *entity.BotEndpoint) error
}

type BotEndpointUsecase struct {
	repo repository.BotEndpointRespositoryInterface
}

func NewBotEndpointUsecase(repo repository.BotEndpointRespositoryInterface) *BotEndpointUsecase {
	return &BotEndpointUsecase{repo: repo}
}

func (usecase *BotEndpointUsecase) RegisterBotEndpoint(botEndpoint *entity.BotEndpoint) error {
	return usecase.repo.Insert(botEndpoint)
}
