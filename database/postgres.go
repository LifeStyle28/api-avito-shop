package database

import (
	"api-avito-shop/models"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

	"crypto/md5"

	_ "github.com/lib/pq"
)

type Postgres struct {
}

func connectToDB() (*sql.DB, error) {
	dbHost := os.Getenv("DATABASE_HOST")
	dbName := os.Getenv("DATABASE_NAME")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbPort := os.Getenv("DATABASE_PORT")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к базе данных: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка при пинге базы данных: %v", err)
	}
	log.Printf("подключение к базе данных успешно установлено\n")
	return db, nil
}

func convertPassToMd5(password string) (string, error) {
	hasher := md5.New()

	_, err := io.WriteString(hasher, password)
	if err != nil {
		return "", fmt.Errorf("ошибка при записи данных в хешер: %w", err)
	}
	hash := hasher.Sum(nil)
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr, nil
}

func (p *Postgres) AddNewUser(username, password string) (bool, error) {
	db, err := connectToDB()
	if err != nil {
		return false, err
	}
	defer db.Close()

	// сначала проверим, что пользователь существует
	var count int
	err = db.QueryRow("SELECT count(*) FROM users WHERE name=$1", username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("ошибка при селекте из базы данных: %w", err)
	}

	// если пользователей нет, то надо создать
	if count == 0 {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return false, fmt.Errorf("старт транзакции: %w", err)
		}
		defer tx.Rollback()

		hashStr, err := convertPassToMd5(password)
		if err != nil {
			return false, err
		}

		var lastInsertId int
		err = tx.QueryRow("INSERT INTO users (name, md5, balance) VALUES($1, $2, $3) RETURNING id", username, hashStr, 1000).Scan(&lastInsertId)
		if err != nil {
			return false, fmt.Errorf("ошибка при добавлении нового пользователя: %w", err)
		}
		log.Println("добавлен новый пользователь с id =", lastInsertId)

		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("ошибка при коммите: %w", err)
		}

		return true, nil
	}
	return false, nil
}

func (p *Postgres) AuthorizeUser(username, password string) (bool, int64, error) {
	db, err := connectToDB()
	if err != nil {
		return false, 0, err
	}
	defer db.Close()

	var id int64
	var md5Pass string
	err = db.QueryRow("SELECT id, md5 FROM users WHERE name=$1", username).Scan(&id, &md5Pass)
	if err != nil {
		return false, 0, fmt.Errorf("ошибка при запросе пароля из базы данных: %v", err)
	}

	hashStr, err := convertPassToMd5(password)
	if err != nil {
		return false, 0, err
	}

	return md5Pass == hashStr, id, nil
}

func (p *Postgres) GetUserCoinsAndItemPrice(userId int64, item string) (float64, float64, int64, error) {
	db, err := connectToDB()
	if err != nil {
		return 0, 0, 0, err
	}
	defer db.Close()

	var coins float64
	err = db.QueryRow("SELECT balance FROM users WHERE id=$1", userId).Scan(&coins)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("ошибка при запросе баланса пользователя из базы данных: %v", err)
	}

	var price float64
	var itemId int64
	err = db.QueryRow("SELECT id, price FROM products WHERE name=$1", item).Scan(&itemId, &price)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("ошибка при запросе стоимости товара из базы данных: %v", err)
	}

	return coins, price, itemId, nil
}

func (p *Postgres) UpdateUserBalanceAndInventory(userId int64, price float64, itemId int64) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("старт транзакции: %w", err)
	}
	defer tx.Rollback()

	// обновим баланс юзера
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id=$1 FOR UPDATE", userId).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь не найден: %d", userId)
		}
		return fmt.Errorf("ошибка при запросе баланса из БД: %w", err)
	}

	newBalance := currentBalance - price
	_, err = tx.Exec("UPDATE users SET balance=$1 WHERE id=$2", newBalance, userId)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении баланса: %w", err)
	}

	// обновим инвентарь юзера
	var quantity int64
	err = tx.QueryRow("SELECT quantity FROM inventory WHERE user_id=$1 AND product_id=$2 FOR UPDATE", userId, itemId).Scan(&quantity)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("ошибка при запросе количества товара из БД: %w", err)
		}
		quantity = 0
	}
	quantity += 1

	_, err = tx.Exec(
		"INSERT INTO inventory (user_id, product_id, quantity) VALUES ($1, $2, 1) ON CONFLICT (user_id, product_id) DO UPDATE SET quantity=$3",
		userId, itemId, quantity)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении инвентаря: %w", err)
	}

	// commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка при коммите: %w", err)
	}

	return nil
}

func (p *Postgres) GetUserCoins(username string) (float64, error) {
	db, err := connectToDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var coins float64
	err = db.QueryRow("SELECT balance FROM users WHERE name=$1", username).Scan(&coins)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка запросе баланса пользователя: %w", err)
	}

	return coins, nil
}

func (p *Postgres) SendCoins(userFrom, userTo string, amount float64) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("старт транзакции: %w", err)
	}
	defer tx.Rollback()

	var userId1, userId2 int64
	var currentBalance float64
	err = tx.QueryRow("SELECT id, balance FROM users WHERE name=$1 FOR UPDATE", userFrom).Scan(&userId1, &currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь не найден: %s", userFrom)
		}
		return fmt.Errorf("ошибка при запросе баланса из БД: %w", err)
	}

	newBalance := currentBalance - amount
	_, err = tx.Exec("UPDATE users SET balance=$1 WHERE name=$2", newBalance, userFrom)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении баланса: %w", err)
	}

	err = tx.QueryRow("SELECT id, balance FROM users WHERE name=$1 FOR UPDATE", userTo).Scan(&userId2, &currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("пользователь не найден: %s", userTo)
		}
		return fmt.Errorf("ошибка при запросе баланса из БД: %w", err)
	}

	newBalance = currentBalance + amount
	_, err = tx.Exec("UPDATE users SET balance=$1 WHERE name=$2", newBalance, userTo)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении баланса: %w", err)
	}

	// запишем транзакцию
	_, err = tx.Exec("INSERT INTO transactions (src, dst, amount) VALUES ($1, $2, $3)", userId1, userId2, amount)
	if err != nil {
		return fmt.Errorf("ошибка при записи транзакции: %w", err)
	}

	// commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка при коммите: %w", err)
	}

	return nil
}

func (p *Postgres) GetUserInventory(userId int64) (*[]models.InfoResponseInventoryInner, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT p.name, i.quantity FROM inventory AS i JOIN products AS p ON p.id = i.product_id WHERE i.user_id = $1", userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer rows.Close()

	goods := make([]models.InfoResponseInventoryInner, 0)
	for rows.Next() {
		var productName string
		var quantity int32
		if err := rows.Scan(&productName, &quantity); err != nil {
			return nil, fmt.Errorf("ошибка получения товаров: %w", err)
		}
		goods = append(goods, models.InfoResponseInventoryInner{Type: productName, Quantity: quantity})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("итерации завершились с ошибкой: %v", err)
	}

	return &goods, nil
}

func (p *Postgres) GetUserReceivedAndSentCoins(userId int64) (*models.InfoResponseCoinHistory, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT u1.name, u2.name, t.amount FROM users AS u1 JOIN transactions AS t ON u1.id=t.src JOIN users AS u2 ON t.dst=u2.id WHERE u1.id=$1", userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer rows.Close()

	sent := make([]models.InfoResponseCoinHistorySentInner, 0)
	for rows.Next() {
		var user1 string
		var user2 string
		var amount float64
		if err := rows.Scan(&user1, &user2, &amount); err != nil {
			return nil, fmt.Errorf("ошибка при получении транзакций отправки: %w", err)
		}
		sent = append(sent, models.InfoResponseCoinHistorySentInner{ToUser: user2, Amount: int32(amount)})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("итерации завершились с ошибкой: %v", err)
	}

	rows, err = db.Query("SELECT u1.name, u2.name, t.amount FROM users AS u1 JOIN transactions AS t ON u1.id=t.dst JOIN users AS u2 ON t.src=u2.id WHERE u1.id=$1", userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer rows.Close()

	received := make([]models.InfoResponseCoinHistoryReceivedInner, 0)
	for rows.Next() {
		var user1 string
		var user2 string
		var amount float64
		if err := rows.Scan(&user1, &user2, &amount); err != nil {
			return nil, fmt.Errorf("ошибка при получении транзакций получения: %w", err)
		}
		received = append(received, models.InfoResponseCoinHistoryReceivedInner{FromUser: user2, Amount: int32(amount)})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("итерации завершились с ошибкой: %v", err)
	}

	history := new(models.InfoResponseCoinHistory)
	history.Received = received
	history.Sent = sent
	return history, nil
}
