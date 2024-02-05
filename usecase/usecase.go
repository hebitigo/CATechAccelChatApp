package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/hebitigo/CATechAccelChatApp/entity"
	"github.com/hebitigo/CATechAccelChatApp/repository"
)

type BotEndpointUsecaseInterface interface {
	RegisterBotEndpoint(dto RegisterBotEndpointInputDTO) error
}

type BotEndpointUsecase struct {
	repo repository.BotEndpointRespositoryInterface
}

func NewBotEndpointUsecase(repo repository.BotEndpointRespositoryInterface) *BotEndpointUsecase {
	return &BotEndpointUsecase{repo: repo}
}

type RegisterBotEndpointInputDTO struct {
	Name     string
	IconURL  string
	Endpoint string
}

func (usecase *BotEndpointUsecase) RegisterBotEndpoint(dto RegisterBotEndpointInputDTO) error {
	botEndpoint := entity.BotEndpoint{Name: dto.Name, IconURL: dto.IconURL, Endpoint: dto.Endpoint}
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
	GetServersByUserID(ctx context.Context, dto GetServersByUserIDInputDTO) ([]entity.Server, error)
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
	var serverId uuid.UUID
	var err error
	err = usecase.TxRepo.DoInTx(ctx, func(ctx context.Context) error {
		server := entity.Server{Name: dto.ServerName}
		serverId, err = usecase.serverRepo.Insert(ctx, server)
		if err != nil {
			return err
		}
		userServer := entity.UserServer{UserId: dto.UserId, ServerId: serverId}
		err = usecase.UserServerRepo.Insert(ctx, userServer)
		if err != nil {
			return err
		}

		channel := entity.Channel{Name: "general", ServerId: serverId}
		_, err = usecase.ChannelRepo.Insert(ctx, channel)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return serverId.String(), nil
}

type GetServersByUserIDInputDTO struct {
	UserId string
}

func (usecase *ServerUsecase) GetServersByUserID(ctx context.Context, dto GetServersByUserIDInputDTO) ([]entity.Server, error) {
	servers, err := usecase.serverRepo.GetServersByUserID(ctx, dto.UserId)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

type UpsertUserInputDTO struct {
	Id           string
	Name         string
	Active       bool
	IconImageURL string
}

type UserUsecaseInterface interface {
	UpsertUser(ctx context.Context, dto UpsertUserInputDTO) error
}

type UserUsecase struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserUsecase(userRepo repository.UserRepositoryInterface) *UserUsecase {
	return &UserUsecase{userRepo: userRepo}
}

func (usecase *UserUsecase) UpsertUser(ctx context.Context, dto UpsertUserInputDTO) error {
	user := entity.User{Id: dto.Id, Name: dto.Name, Active: dto.Active, IconImageURL: dto.IconImageURL}
	err := usecase.userRepo.Upsert(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

type ChannelUsecaseInterface interface {
	GetChannelsByServerID(ctx context.Context, dto GetChannelsByServerIDInputDTO) ([]entity.Channel, error)
	RegisterChannel(ctx context.Context, dto RegisterChannelInputDTO) (string, error)
}

type ChannelUsecase struct {
	channelRepo repository.ChannelRepositoryInterface
}

func NewChannelUsecase(channelRepo repository.ChannelRepositoryInterface) *ChannelUsecase {
	return &ChannelUsecase{channelRepo: channelRepo}
}

// UserId以外のIdを元にデータを取得する際は、entityのIdの型を参考にしてInputDTOのIdの型を決める
// serverIdはentityではServer構造体のIdの型は*uuid.UUIDだけど、これはデータベース側にuuidを自動生成させる
// ためにポインタを使ってServer構造体のIdの値をnilにできるようにしているだけで、実際にIdをもとに検索を書ける場合は
// Idの値がnilになることはないので、InputDTOのServerIdの型はuuid.UUIDになる
type GetChannelsByServerIDInputDTO struct {
	ServerId uuid.UUID
}

func (usecase *ChannelUsecase) GetChannelsByServerID(ctx context.Context, dto GetChannelsByServerIDInputDTO) ([]entity.Channel, error) {
	channels, err := usecase.channelRepo.GetChannelsByServerID(ctx, dto.ServerId)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

type RegisterChannelInputDTO struct {
	ServerId    uuid.UUID
	ChannelName string
}

func (usecase *ChannelUsecase) RegisterChannel(ctx context.Context, dto RegisterChannelInputDTO) (string, error) {
	channel := entity.Channel{Name: dto.ChannelName, ServerId: dto.ServerId}
	channelId, err := usecase.channelRepo.Insert(ctx, channel)
	if err != nil {
		return "", err
	}
	return channelId.String(), nil
}
