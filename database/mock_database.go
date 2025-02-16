package database

import (
	"api-avito-shop/models"
	"fmt"

	"github.com/stretchr/testify/mock"
)

type User struct {
	id        int64
	username  string
	password  string
	balance   float64
	inventory []models.InfoResponseInventoryInner
	history   models.InfoResponseCoinHistory
}

type MockDatabase struct {
	mock.Mock
	users    []User
	usersMap map[string]string
	nextId   int64
}

var ProductsMap map[string]struct {
	itemId int64
	Price  int64
}

func init() {
	ProductsMap = make(map[string]struct {
		itemId int64
		Price  int64
	})

	ProductsMap["t-shirt"] = struct {
		itemId int64
		Price  int64
	}{itemId: 1, Price: 100}

	ProductsMap["cup"] = struct {
		itemId int64
		Price  int64
	}{itemId: 2, Price: 20}

	ProductsMap["something_else"] = struct {
		itemId int64
		Price  int64
	}{itemId: 3, Price: 30}
}

func NewMockDb() *MockDatabase {
	return &MockDatabase{
		usersMap: make(map[string]string),
	}
}

const AddUserKey = "add_user"
const AuthorizeUserKey = "authorize_user"
const GetUserCoinsKey = "user_coins"
const UpdateUserBalanceAndInventoryKey = "update_user_balance_inventory"
const UserCoinsAndItemPriceKey = "user_coins_and_item_price"
const UserInventoryKey = "user_inventory"
const UserTransactionsKey = "user_transactions"
const SendCoinsKey = "send_coins"

func (m *MockDatabase) ErrorWithDb(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockDatabase) checkUserById(userId int64) bool {
	for _, user := range m.users {
		if user.id == userId {
			return true
		}
	}
	return false
}

func (m *MockDatabase) getUserIdByUserName(username string) int64 {
	for _, user := range m.users {
		if user.username == username {
			return user.id
		}
	}
	return 0
}

func (m *MockDatabase) checkUserByUsername(username string) bool {
	_, ok := m.usersMap[username]
	return ok
}

func (m *MockDatabase) AddNewUser(username, password string) (bool, error) {
	err := m.ErrorWithDb(AddUserKey)
	if err != nil {
		return false, err
	}

	if _, ok := m.usersMap[username]; ok {
		return false, nil
	}
	m.usersMap[username] = password
	m.users = append(m.users, User{id: m.nextId, username: username, password: password, balance: 1000})
	m.nextId += 1
	return true, nil
}

func (m *MockDatabase) AuthorizeUser(username, password string) (bool, int64, error) {
	err := m.ErrorWithDb(AuthorizeUserKey)
	if err != nil {
		return false, 0, err
	}

	if _, ok := m.usersMap[username]; !ok {
		return false, 0, nil
	}

	for _, user := range m.users {
		if user.username == username && user.password == password {
			return true, user.id, nil
		}
	}
	return false, 0, nil
}

func (m *MockDatabase) GetUserCoinsAndItemPrice(userId int64, item string) (float64, float64, int64, error) {
	err := m.ErrorWithDb(UserCoinsAndItemPriceKey)
	if err != nil {
		return 0, 0, 0, err
	}

	if _, ok := ProductsMap[item]; !ok {
		return 0, 0, 0, fmt.Errorf("no product in product map")
	}

	ok := false
	for _, user := range m.users {
		if user.id == userId {
			ok = true
		}
	}

	if !ok {
		return 0, 0, 0, fmt.Errorf("user not found")
	}

	return m.users[userId].balance, float64(ProductsMap[item].Price), ProductsMap[item].itemId, nil
}

func (m *MockDatabase) UpdateUserBalanceAndInventory(userId int64, price float64, itemId int64) error {
	err := m.ErrorWithDb(UpdateUserBalanceAndInventoryKey)
	if err != nil {
		return err
	}

	ok := false
	var typeKey string
	for key, item := range ProductsMap {
		if item.itemId == itemId {
			ok = true
			typeKey = key
		}
	}

	if !ok {
		return fmt.Errorf("no product in product map")
	}

	m.users[userId].balance -= price
	for i, good := range m.users[userId].inventory {
		if _, ok := ProductsMap[good.Type]; ok {
			m.users[userId].inventory[i].Quantity += 1
			return nil
		}
	}
	m.users[userId].inventory = append(m.users[userId].inventory, models.InfoResponseInventoryInner{Type: typeKey, Quantity: 1})
	return nil
}

func (m *MockDatabase) GetUserCoins(username string) (float64, error) {
	err := m.ErrorWithDb(GetUserCoinsKey)
	if err != nil {
		return 0, err
	}

	if _, ok := m.usersMap[username]; !ok {
		// не должно стрелять, т.к. в движке проверяется изначально, что юзер есть
		return 0, fmt.Errorf("not enough user in db")
	}

	for _, user := range m.users {
		if user.username == username {
			return user.balance, nil
		}
	}

	// не должно стрелять, т.к. в движке проверяется изначально, что юзер есть
	return 0, fmt.Errorf("not enough user in db")
}

func (m *MockDatabase) SendCoins(userFrom, userTo string, amount float64) error {
	err := m.ErrorWithDb(SendCoinsKey)
	if err != nil {
		return err
	}

	if !m.checkUserByUsername(userFrom) || !m.checkUserByUsername(userTo) {
		return fmt.Errorf("not enough user in db")
	}

	userFromId := m.getUserIdByUserName(userFrom)
	userToId := m.getUserIdByUserName(userTo)

	m.users[userFromId].balance -= amount
	m.users[userToId].balance += amount

	m.users[userFromId].history.Sent = append(m.users[userFromId].history.Sent,
		models.InfoResponseCoinHistorySentInner{ToUser: userTo, Amount: int32(amount)})

	m.users[userToId].history.Received = append(m.users[userToId].history.Received,
		models.InfoResponseCoinHistoryReceivedInner{FromUser: userFrom, Amount: int32(amount)})

	return nil
}

func (m *MockDatabase) GetUserInventory(userId int64) (*[]models.InfoResponseInventoryInner, error) {
	err := m.ErrorWithDb(UserInventoryKey)
	if err != nil {
		return nil, err
	}

	if !m.checkUserById(userId) {
		return nil, fmt.Errorf("user not found error")
	}

	return &m.users[userId].inventory, nil
}

func (m *MockDatabase) GetUserReceivedAndSentCoins(userId int64) (*models.InfoResponseCoinHistory, error) {
	err := m.ErrorWithDb(UserTransactionsKey)
	if err != nil {
		return nil, err
	}

	if !m.checkUserById(userId) {
		return nil, fmt.Errorf("user not found")
	}

	return &m.users[userId].history, nil
}
