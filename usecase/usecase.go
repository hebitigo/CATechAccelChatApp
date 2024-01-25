package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hebitigo/CATechAccelChatApp/entity"
	"github.com/hebitigo/CATechAccelChatApp/repository"
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

// TODO:Endpointのrequestの構造体をhandler側に記述して、それをusecase側で受け取るようにする
// requestに対応する構造体からentity.BotEndpointに変換する処理を書く
// 返答用の構造体もhandler側に書く
func (usecase *BotEndpointUsecase) RegisterBotEndpoint(botEndpoint *entity.BotEndpoint) error {
	return usecase.repo.Insert(botEndpoint)
}

func RegisterMessage() {
	//wsパッケージの処理から受け取った
	//channel経由でメッセージを受け取ってDBに登録する処理をメッセージのIsBotで判断して
	//repositoryから処理をinterface経由で引用して書く
}

func RegisterUser() {
	//userがログインした際に登録処理をrepository経由で走らせる
}

type ServerUsecaseInterface interface {
	RegisterServer(ctx context.Context, dto RegisterServerInputDTO) (string, error)
}

type ServerUsecase struct {
	serverRepo     repository.ServerRepositoryInterface
	ChannelRepo    repository.ChannelRepositoryInterface
	UserServerRepo repository.UserServerRepositoryInterface
	TxRepo         repository.TxRepositoryInterface
}

func NewServerUsecase(serverRepo repository.ServerRepositoryInterface, channelRepo repository.ChannelRepositoryInterface, userServerRepo repository.UserServerRepositoryInterface, txRepo repository.TxRepositoryInterface) *ServerUsecase {
	return &ServerUsecase{serverRepo: serverRepo, ChannelRepo: channelRepo, UserServerRepo: userServerRepo, TxRepo: txRepo}
}

type RegisterServerInputDTO struct {
	ServerName string
	UserId     string
}

func (usecase *ServerUsecase) RegisterServer(ctx context.Context, dto RegisterServerInputDTO) (string, error) {
	var serverId *uuid.UUID
	var err error
	err = usecase.TxRepo.DoInTx(ctx, func(ctx context.Context) error {
		server := entity.Server{Name: dto.ServerName}
		serverId, err = usecase.serverRepo.Insert(ctx, server)
		if err != nil {
			return err
		}
		userServer := entity.UserServer{UserId: dto.UserId, ServerId: serverId}
		if err := usecase.UserServerRepo.Insert(ctx, userServer); err != nil {
			return err
		}

		channel := entity.Channel{Name: "default", ServerId: serverId}
		if err := usecase.ChannelRepo.Insert(ctx, channel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return serverId.String(), nil
}

type RegisterUserInputDTO struct {
	Id           string
	Name         string
	Active       bool
	IconImageURL string
}

type UserUsecaseInterface interface {
	RegisterUser(ctx context.Context, dto RegisterUserInputDTO) error
}

type UserUsecase struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserUsecase(userRepo repository.UserRepositoryInterface) *UserUsecase {
	return &UserUsecase{userRepo: userRepo}
}

func (usecase *UserUsecase) RegisterUser(ctx context.Context, dto RegisterUserInputDTO) error {
	user := entity.User{Id: dto.Id, Name: dto.Name, Active: dto.Active, IconImageURL: dto.IconImageURL}
	err := usecase.userRepo.Insert(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
