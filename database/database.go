package database

import "api-avito-shop/models"

type Database interface {
	AddNewUser(username, password string) (bool, error)
	AuthorizeUser(username, password string) (bool, int64, error)
	GetUserCoinsAndItemPrice(userId int64, item string) (float64, float64, int64, error)
	UpdateUserBalanceAndInventory(userId int64, price float64, itemId int64) error
	GetUserCoins(username string) (float64, error)
	SendCoins(userFrom, userTo string, amount float64) error
	GetUserInventory(userId int64) (*[]models.InfoResponseInventoryInner, error)
	GetUserReceivedAndSentCoins(userId int64) (*models.InfoResponseCoinHistory, error)
}
