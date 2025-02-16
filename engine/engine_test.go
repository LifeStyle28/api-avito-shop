package engine

import (
	"api-avito-shop/database"
	"api-avito-shop/models"
	"context"
	"errors"
	"testing"

	"github.com/form3tech-oss/jwt-go"
	"github.com/stretchr/testify/assert"
)

func addTokenToCtx(ctx *context.Context, tokenString string) {
	parsedToken, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(models.JwtUniqueKey), nil
	})
	*ctx = context.WithValue(*ctx, models.JwtUserKey, parsedToken)
}

func TestHandleApiAuthWitAddNewUser(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// добавление нового пользователя
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	coins, _ := mockDb.GetUserCoins(username)
	assert.True(t, coins == float64(1000))

	// теперь попробуем позвать добавленного пользователя с другим паролем
	password = "some_pass"
	req = models.AuthRequest{Username: username, Password: password}
	resp, _ = e.HandleApiAuth(ctx, req)
	assert.True(t, int(401) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorPassword} == resp.Body)
}

func TestHandleApiAuthWitAddNewUserErrorDb(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// добавление нового пользователя, имитируем ошибку БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(errors.New("error"))
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorAddNewUser} == resp.Body)
}

func TestHandleApiAuthWitAutorizeUserErrorDb(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(errors.New("error"))
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)

	// теперь пробуем авторизоваться заново, чтобы словить ошибку БД
	req = models.AuthRequest{Username: username, Password: password}
	resp, _ = e.HandleApiAuth(ctx, req)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorUserAuthorize} == resp.Body)
}

func TestHandleApiBuyItem(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UpdateUserBalanceAndInventoryKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserCoinsAndItemPriceKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserInventoryKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserTransactionsKey).Return(nil)

	// cоздадим юзера
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx, resp.Body.(models.AuthResponse).Token)

	// запросим инфо по пользователю и скопируем баланс, чтобы потом проверить
	resp, _ = e.HandleApiInfo(ctx)
	assert.True(t, int(200) == resp.Code)
	oldBalance := resp.Body.(models.InfoResponse).Coins

	// купим футболку
	resp, _ = e.HandleApiByuItem(ctx, "t-shirt")
	assert.True(t, int(200) == resp.Code)

	// проверим, что у юзера стал меньше баланс на стоимость футболки
	resp, _ = e.HandleApiInfo(ctx)
	assert.True(t, int(200) == resp.Code)
	newBalance := resp.Body.(models.InfoResponse).Coins
	assert.True(t, int64(oldBalance-newBalance) == database.ProductsMap["t-shirt"].Price)

	// проверим, что у юзера всего 1 футболка
	assert.True(t, resp.Body.(models.InfoResponse).Inventory[0].Quantity == 1)
}

func TestHandleApiBuyItemErrorDbWithUserCoins(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что будет ошибка получения коинов
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(errors.New("error"))

	// cоздадим юзера
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx, resp.Body.(models.AuthResponse).Token)

	// проверим, что вернется ошибка 500
	resp, _ = e.HandleApiInfo(ctx)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorUserData + username} == resp.Body)
}

func TestHandleApiBuyItemErrorDbWithUserInventory(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что будет ошибка получения инвентаря
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserInventoryKey).Return(errors.New("error"))

	// cоздадим юзера
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx, resp.Body.(models.AuthResponse).Token)

	// проверим, что вернется ошибка 500
	resp, _ = e.HandleApiInfo(ctx)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorInventory} == resp.Body)
}

func TestHandleApiBuyItemErrorDbWithUserTransactions(t *testing.T) {
	username := "test_user1"
	password := "test_pass1"

	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что будет ошибка получения транзакций пользователя
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserInventoryKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserTransactionsKey).Return(errors.New("error"))

	// cоздадим юзера
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx, resp.Body.(models.AuthResponse).Token)

	// проверим, что вернется ошибка 500
	resp, _ = e.HandleApiInfo(ctx)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorTransactions} == resp.Body)
}

func TestHandleApiSendCoin(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.SendCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserInventoryKey).Return(nil)
	mockDb.On("ErrorWithDb", database.UserTransactionsKey).Return(nil)

	// cоздадим первого юзера
	username1 := "test_user1"
	password1 := "test_pass1"
	req := models.AuthRequest{Username: username1, Password: password1}
	resp, _ := e.HandleApiAuth(ctx1, req)
	assert.True(t, int(200) == resp.Code)

	// сохраним баланс первого пользователя
	addTokenToCtx(&ctx1, resp.Body.(models.AuthResponse).Token)
	resp, _ = e.HandleApiInfo(ctx1)
	assert.True(t, int(200) == resp.Code)
	oldBalanceUser1 := resp.Body.(models.InfoResponse).Coins

	// cоздадим второго юзера
	username2 := "test_user2"
	password2 := "test_pass2"
	req = models.AuthRequest{Username: username2, Password: password2}
	resp, _ = e.HandleApiAuth(ctx2, req)
	assert.True(t, int(200) == resp.Code)

	// сохраним баланс второго пользователя
	addTokenToCtx(&ctx2, resp.Body.(models.AuthResponse).Token)
	resp, _ = e.HandleApiInfo(ctx1)
	assert.True(t, int(200) == resp.Code)
	oldBalanceUser2 := resp.Body.(models.InfoResponse).Coins

	// перешлём коины от первого юзера второму
	amount := int32(200)
	reqSendCoins := models.SendCoinRequest{ToUser: username2, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx1, reqSendCoins)
	assert.True(t, int(200) == resp.Code)

	// проверим что у первого юзера есть в истории данная транзакция и его монеты уменьшились на amount
	resp, _ = e.HandleApiInfo(ctx1)
	assert.True(t, int(200) == resp.Code)
	toUser := resp.Body.(models.InfoResponse).CoinHistory.Sent[0].ToUser
	toAmount := resp.Body.(models.InfoResponse).CoinHistory.Sent[0].Amount
	assert.True(t, toUser == username2)
	assert.True(t, toAmount == amount)
	newBalanceUser1 := resp.Body.(models.InfoResponse).Coins
	assert.True(t, oldBalanceUser1-newBalanceUser1 == amount)

	// проверим что у второго юзера есть в истории данная транзакция и его монеты увеличились на amount
	resp, _ = e.HandleApiInfo(ctx2)
	assert.True(t, int(200) == resp.Code)
	fromUser := resp.Body.(models.InfoResponse).CoinHistory.Received[0].FromUser
	fromAmount := resp.Body.(models.InfoResponse).CoinHistory.Received[0].Amount
	assert.True(t, fromUser == username1)
	assert.True(t, fromAmount == amount)
	newBalanceUser2 := resp.Body.(models.InfoResponse).Coins
	assert.True(t, newBalanceUser2-oldBalanceUser2 == amount)
}

func TestHandleApiSendCoinSameUser(t *testing.T) {
	ctx := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)

	// cоздадим юзера
	username := "test_user1"
	password := "test_pass1"
	req := models.AuthRequest{Username: username, Password: password}
	resp, _ := e.HandleApiAuth(ctx, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx, resp.Body.(models.AuthResponse).Token)

	// пытаемся послать коины самому себе
	amount := int32(200)
	reqSendCoins := models.SendCoinRequest{ToUser: username, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx, reqSendCoins)
	assert.True(t, int(400) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorSameUser} == resp.Body)
}

func TestHandleApiSendCoinErrorDbWithUserCoinsWithUser1(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(errors.New("error"))

	// cоздадим первого юзера
	username1 := "test_user1"
	password1 := "test_pass1"
	req := models.AuthRequest{Username: username1, Password: password1}
	resp, _ := e.HandleApiAuth(ctx1, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx1, resp.Body.(models.AuthResponse).Token)

	// cоздадим второго юзера
	username2 := "test_user2"
	password2 := "test_pass2"
	req = models.AuthRequest{Username: username2, Password: password2}
	resp, _ = e.HandleApiAuth(ctx2, req)
	assert.True(t, int(200) == resp.Code)

	// перешлём коины от первого юзера второму
	amount := int32(200)
	reqSendCoins := models.SendCoinRequest{ToUser: username2, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx1, reqSendCoins)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorUserData + username1} == resp.Body)
}

func TestHandleApiSendCoinErrorDbWithUserCoinsWithUser2(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil).Once()
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(errors.New("error"))

	// cоздадим первого юзера
	username1 := "test_user1"
	password1 := "test_pass1"
	req := models.AuthRequest{Username: username1, Password: password1}
	resp, _ := e.HandleApiAuth(ctx1, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx1, resp.Body.(models.AuthResponse).Token)

	// cоздадим второго юзера
	username2 := "test_user2"
	password2 := "test_pass2"
	req = models.AuthRequest{Username: username2, Password: password2}
	resp, _ = e.HandleApiAuth(ctx2, req)
	assert.True(t, int(200) == resp.Code)

	// перешлём коины от первого юзера второму
	amount := int32(200)
	reqSendCoins := models.SendCoinRequest{ToUser: username2, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx1, reqSendCoins)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorUserData + username2} == resp.Body)
}

func TestHandleApiSendCoinAmountGreateBalance(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)

	// cоздадим первого юзера
	username1 := "test_user1"
	password1 := "test_pass1"
	req := models.AuthRequest{Username: username1, Password: password1}
	resp, _ := e.HandleApiAuth(ctx1, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx1, resp.Body.(models.AuthResponse).Token)

	// cоздадим второго юзера
	username2 := "test_user2"
	password2 := "test_pass2"
	req = models.AuthRequest{Username: username2, Password: password2}
	resp, _ = e.HandleApiAuth(ctx2, req)
	assert.True(t, int(200) == resp.Code)

	// перешлём коины от первого юзера второму
	amount := int32(1200)
	reqSendCoins := models.SendCoinRequest{ToUser: username2, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx1, reqSendCoins)
	assert.True(t, int(400) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorUserBalance} == resp.Body)
}

func TestHandleApiSendCoinErrorDbWithSendCoins(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	mockDb := database.NewMockDb()
	e := NewEngine(mockDb)

	// мокируем, что не будет ошибок БД
	mockDb.On("ErrorWithDb", database.AddUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.AuthorizeUserKey).Return(nil)
	mockDb.On("ErrorWithDb", database.GetUserCoinsKey).Return(nil)
	mockDb.On("ErrorWithDb", database.SendCoinsKey).Return(errors.New("error"))

	// cоздадим первого юзера
	username1 := "test_user1"
	password1 := "test_pass1"
	req := models.AuthRequest{Username: username1, Password: password1}
	resp, _ := e.HandleApiAuth(ctx1, req)
	assert.True(t, int(200) == resp.Code)
	addTokenToCtx(&ctx1, resp.Body.(models.AuthResponse).Token)

	// cоздадим второго юзера
	username2 := "test_user2"
	password2 := "test_pass2"
	req = models.AuthRequest{Username: username2, Password: password2}
	resp, _ = e.HandleApiAuth(ctx2, req)
	assert.True(t, int(200) == resp.Code)

	// перешлём коины от первого юзера второму
	amount := int32(200)
	reqSendCoins := models.SendCoinRequest{ToUser: username2, Amount: amount}
	resp, _ = e.HandleApiSendCoin(ctx1, reqSendCoins)
	assert.True(t, int(500) == resp.Code)
	assert.True(t, models.ErrorResponse{Errors: ErrorSendCoin} == resp.Body)
}
