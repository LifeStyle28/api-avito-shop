package engine

import (
	"api-avito-shop/database"
	"api-avito-shop/models"
	"context"
	"time"

	"github.com/form3tech-oss/jwt-go"
)

type Engine struct {
	db database.Database
}

type AccountData struct {
	Id       int64
	Username string
	Password string
}

func NewEngine(db database.Database) *Engine {
	return &Engine{
		db: db,
	}
}

func extractTokenFromContext(ctx context.Context) *jwt.Token {
	if ctx == nil {
		return nil
	}

	tokenInterface := ctx.Value(models.JwtUserKey)
	if tokenInterface == nil {
		return nil
	}

	token, ok := tokenInterface.(*jwt.Token)
	if !ok {
		return nil
	}

	return token
}

func (e *Engine) getAccountData(ctx context.Context) (*AccountData, models.ImplResponse) {
	token := extractTokenFromContext(ctx).Claims.(jwt.MapClaims)
	password, ok := token["password"].(string)
	if !ok {
		return nil, models.Response(500, models.ErrorResponse{Errors: ErrorPasswordType})
	}
	username, ok := token["username"].(string)
	if !ok {
		return nil, models.Response(500, models.ErrorResponse{Errors: ErrorUserName})
	}

	isAuthorize, userId, err := e.db.AuthorizeUser(username, password)
	if err != nil {
		return nil, models.Response(500, models.ErrorResponse{Errors: ErrorUserAuthorize})
	}
	if !isAuthorize {
		return nil, models.Response(401, models.ErrorResponse{Errors: ErrorPassword})
	}

	data := new(AccountData)
	data.Id = userId
	data.Password = password
	data.Username = username

	return data, models.ImplResponse{}
}

func (e *Engine) HandleApiInfo(ctx context.Context) (models.ImplResponse, error) {
	data, response := e.getAccountData(ctx)
	if data == nil {
		return response, nil
	}
	coins, err := e.db.GetUserCoins(data.Username)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorUserData + data.Username}), nil
	}

	goods, err := e.db.GetUserInventory(data.Id)
	if err != nil || goods == nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorInventory}), nil
	}
	history, err := e.db.GetUserReceivedAndSentCoins(data.Id)
	if err != nil || history == nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorTransactions}), nil
	}
	return models.Response(200, models.InfoResponse{Coins: int32(coins), Inventory: *goods, CoinHistory: *history}), nil
}

func (e *Engine) HandleApiSendCoin(ctx context.Context, sendCoinRequest models.SendCoinRequest) (models.ImplResponse, error) {
	data, response := e.getAccountData(ctx)
	if data == nil {
		return response, nil
	}

	if data.Username == sendCoinRequest.ToUser {
		return models.Response(400, models.ErrorResponse{Errors: ErrorSameUser}), nil
	}

	coinsFrom, err := e.db.GetUserCoins(data.Username)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorUserData + data.Username}), nil
	}

	_, err = e.db.GetUserCoins(sendCoinRequest.ToUser)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorUserData + sendCoinRequest.ToUser}), nil
	}

	if float64(sendCoinRequest.Amount) > coinsFrom {
		return models.Response(400, models.ErrorResponse{Errors: ErrorUserBalance}), nil
	}

	err = e.db.SendCoins(data.Username, sendCoinRequest.ToUser, float64(sendCoinRequest.Amount))
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorSendCoin}), nil
	}
	return models.Response(200, models.ImplResponse{}), nil
}

func (e *Engine) HandleApiByuItem(ctx context.Context, item string) (models.ImplResponse, error) {
	data, response := e.getAccountData(ctx)
	if data == nil {
		return response, nil
	}

	coins, price, itemId, err := e.db.GetUserCoinsAndItemPrice(data.Id, item)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorDatabase}), nil
	}

	if coins < price {
		return models.Response(400, models.ErrorResponse{Errors: ErrorUserBalance}), nil
	}

	err = e.db.UpdateUserBalanceAndInventory(data.Id, price, itemId)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorUpdateUserBalance}), nil
	}

	return models.Response(200, models.ImplResponse{}), nil
}

func (e *Engine) HandleApiAuth(ctx context.Context, authRequest models.AuthRequest) (models.ImplResponse, error) {
	var key = []byte(models.JwtUniqueKey)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["password"] = authRequest.Password
	claims["username"] = authRequest.Username
	claims["admin"] = false
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Токен истекает через 24 часа

	tokenString, err := token.SignedString(key)
	if err != nil {
		return models.Response(500, models.ErrorResponse{}), nil
	}
	isAdd, err := e.db.AddNewUser(authRequest.Username, authRequest.Password)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorAddNewUser}), nil
	}
	if isAdd {
		return models.Response(200, models.AuthResponse{Token: tokenString}), nil
	}

	isAuthorize, _, err := e.db.AuthorizeUser(authRequest.Username, authRequest.Password)
	if err != nil {
		return models.Response(500, models.ErrorResponse{Errors: ErrorUserAuthorize}), nil
	}
	if !isAuthorize {
		return models.Response(401, models.ErrorResponse{Errors: ErrorPassword}), nil
	}
	return models.Response(200, models.AuthResponse{Token: tokenString}), nil
}
