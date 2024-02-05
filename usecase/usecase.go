package usecase

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

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
	CreateInvitationByJWT(dto CreateInvitationByJWTInputDTO) ([]byte, error)
	AuthAndAddUser(ctx context.Context, dto AuthAndAddUserInputDTO) (*entity.Server, error)
}

type ServerUsecase struct {
	serverRepo     repository.ServerRepositoryInterface
	channelRepo    repository.ChannelRepositoryInterface
	userServerRepo repository.UserServerRepositoryInterface
	userRepo       repository.UserRepositoryInterface
	txRepo         repository.TxRepositoryInterface
}

func NewServerUsecase(serverRepo repository.ServerRepositoryInterface, channelRepo repository.ChannelRepositoryInterface, userServerRepo repository.UserServerRepositoryInterface, txRepo repository.TxRepositoryInterface, userRepo repository.UserRepositoryInterface) *ServerUsecase {
	return &ServerUsecase{serverRepo: serverRepo, channelRepo: channelRepo, userServerRepo: userServerRepo, txRepo: txRepo, userRepo: userRepo}
}

type CreateInvitationByJWTInputDTO struct {
	ServerId string
	UserId   string
}

func (usecase *ServerUsecase) CreateInvitationByJWT(dto CreateInvitationByJWTInputDTO) ([]byte, error) {
	//jwtを生成する処理を書く
	token, err := jwt.NewBuilder().Claim("serverId", dto.ServerId).Claim("issuerId", dto.UserId).IssuedAt(time.Now()).Expiration(time.Now().Add(time.Minute * 30)).Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build token")
	}
	key, err := os.ReadFile("secret.pem")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read key file")
	}
	secKey, err := jwk.ParseKey(key, jwk.WithPEM(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key")
	}
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, secKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign token")
	}
	return signed, nil
}

type AuthAndAddUserInputDTO struct {
	Token  []byte
	UserId string
}

func (usecase *ServerUsecase) AuthAndAddUser(ctx context.Context, dto AuthAndAddUserInputDTO) (*entity.Server, error) {
	//参考になりそう
	//https://github.com/lestrrat-go/jwx/blob/d86010aad62ff60ad593f97f39c2ea3e8ab5691e/examples/jwt_example_test.go#L166C1-L169C73
	//https://github.com/lestrrat-go/jwx/blob/d86010aad62ff60ad593f97f39c2ea3e8ab5691e/examples/jwt_example_test.go#L80
	key, err := os.ReadFile("public.pem")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read key file")
	}
	pubKey, err := jwk.ParseKey(key, jwk.WithPEM(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key")
	}

	//https://pkg.go.dev/github.com/lestrrat-go/jwx/v2@v2.0.19/jwt#Parse
	//If the token is signed and you want to verify the payload matches the signature, you must pass the jwt.WithKey(alg, key) or jwt.WithKeySet(jwk.Set) option. If you do not specify these parameters, no verification will be performed.
	//>トークンが署名されていて、ペイロードが署名と一致することを検証したい場合は、jwt.WithKey(alg, key)またはjwt.WithKeySet(jwk.Set)オプションを渡す必要があります。これらのパラメータを指定しない場合、検証は行われません。

	//If you also want to assert the validity of the JWT itself (i.e. expiration and such), use the `Validate()` function on the returned token, or pass the `WithValidate(true)` option. Validate options can also be passed to `Parse`
	//This function takes both ParseOption and ValidateOption types: ParseOptions control the parsing behavior, and ValidateOptions are passed to `Validate()` when `jwt.WithValidate` is specified.
	//>JWT 自体の有効性（有効期限など）も保証したい場合は、返されたトークンに `Validate()` 関数を使うか、`WithValidate(true)` オプションを渡す。Validate オプションは `Parse` にも渡すことができる。
	//>この関数は ParseOption 型と ValidateOption 型の両方を受け取ります：ParseOptionsはパースの動作を制御し、ValidateOptionsは `jwt.WithValidate` が指定されたときに `Validate()` に渡されます。

	//payload, err := jws.Verify(dto.Token, jws.WithKey(jwa.RS256, pubKey))
	//だとexpクレームが有効期限切れてるのに何故か認証が通ったが、jwt.Parseを使うと有効期限切れの場合はエラーが返る
	//
	//https://pkg.go.dev/github.com/lestrrat-go/jwx/v2@v2.0.19/jwt#WithValidate
	// WithValidate is passed to `Parse()` method to denote that the validation of the JWT token should be performed (or not) after a successful parsing of the incoming payload.
	// This option is enabled by default.
	// If you would like disable validation, you must use `jwt.WithValidate(false)` or use `jwt.ParseInsecure()`
	//>WithValidate は、受信したペイロードのパースが成功した後に JWT トークンの検証を実行する（または実行しない）ことを示すために `Parse()` メソッドに渡されます。
	//>"このオプションはデフォルトで有効になっています"。
	//↑多分これのおかげでexpクレームが有効期限切れの場合はエラーが返る
	// {
	//   "error": "failed to verify jwt: \"exp\" not satisfied"
	// }
	payload, err := jwt.Parse(dto.Token, jwt.WithKey(jwa.RS256, pubKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify jwt")
	}

	log.Printf("payload: %v", payload)

	claims := payload.PrivateClaims()

	serverId, ok := claims["serverId"].(string)
	if !ok {
		return nil, errors.New("failed to get serverId from jwt")
	}
	issuerId, ok := claims["issuerId"].(string)
	if !ok {
		return nil, errors.New("failed to get userId from jwt")
	}
	err = usecase.userRepo.CheckUserExist(ctx, issuerId)
	if err != nil {
		return nil, err
	}
	//serverIdをUUIDに変換
	serverUUID, err := uuid.Parse(serverId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse serverId")
	}
	userServer := entity.UserServer{UserId: dto.UserId, ServerId: &serverUUID}
	err = usecase.userServerRepo.Insert(ctx, userServer)
	if err != nil {
		return nil, err
	}
	server, err := usecase.serverRepo.GetServer(ctx, serverId)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

type RegisterServerInputDTO struct {
	ServerName string
	UserId     string
}

func (usecase *ServerUsecase) RegisterServer(ctx context.Context, dto RegisterServerInputDTO) (string, error) {
	var serverId *uuid.UUID
	var err error
	err = usecase.txRepo.DoInTx(ctx, func(ctx context.Context) error {
		server := entity.Server{Name: dto.ServerName}
		serverId, err = usecase.serverRepo.Insert(ctx, server)
		if err != nil {
			return err
		}
		userServer := entity.UserServer{UserId: dto.UserId, ServerId: serverId}
		err = usecase.userServerRepo.Insert(ctx, userServer)
		if err != nil {
			return err
		}

		channel := entity.Channel{Name: "default", ServerId: serverId}
		err = usecase.channelRepo.Insert(ctx, channel)
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
