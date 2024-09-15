package database

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBHandler struct {
	db  *sql.DB
	ctx context.Context
}
type OrderGetJSON struct {
	Order        string  `json:"number"`
	Status       string  `json:"status"`
	Accrual      float64 `json:"accrual,omitempty"`
	Processed_at string  `json:"uploaded_at"`
}

func InitDB(DNS string, ctx context.Context) (*DBHandler, error) {
	db, err := sql.Open("pgx", DNS)
	if err != nil {
		log.Println(err.Error())
	}
	dbh := &DBHandler{
		db:  db,
		ctx: ctx,
	}
	return dbh, nil
}

func CreateChema(db *DBHandler) error {
	createTableUsers :=
		`CREATE TABLE IF NOT EXISTS users(
			login VARCHAR(200) PRIMARY KEY,
			password VARCHAR(200) NOT NULL,
    		user_id VARCHAR(200) UNIQUE
		)`

	createTableOrders :=
		`CREATE TABLE IF NOT EXISTS orders(
			order_id VARCHAR(200) PRIMARY KEY,
			user_id VARCHAR(200) NOT NULL,
			accrual  float,
			status VARCHAR(200),
			uploadet_at timestamp default current_timestamp
		)`

	_, err := db.db.Exec(createTableUsers)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(createTableOrders)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) CheckDuplicateLogin(login string) (bool, error) {
	var double string
	err := handler.db.QueryRow(
		"SELECT login FROM users WHERE login =$1", login).Scan(&double)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return false, nil
		}
		log.Println(err.Error())
		return true, err
	}
	if double != "" {
		return true, nil
	}
	return false, nil
}

func (handler DBHandler) RegisterUser(login string, pass string, cookie string) error {

	_, err := handler.db.Exec("INSERT INTO users (login, password, user_id) VALUES($1,$2,$3)", login, pass, cookie)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) GetUserPassword(login string) (string, error) {
	var password string

	err := handler.db.QueryRow(
		"SELECT password FROM users where login =$1", login).Scan(&password)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	if password != "" {
		return password, nil
	}
	return "", nil
}

func (handler DBHandler) GetUserID(login string) (string, error) {
	var ID string

	err := handler.db.QueryRow(
		"SELECT user_id FROM users where login =$1", login).Scan(&ID)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return ID, nil
}

func (handler DBHandler) GetUseIDByOrderID(order string) (string, error) {
	var ID string

	err := handler.db.QueryRow(
		"SELECT user_id FROM orders where order_id =$1", order).Scan(&ID)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return "", nil
		}
		log.Println(err.Error())
		return "", err
	}
	return ID, nil
}

func (handler DBHandler) GetOrdersByUserID(userID string) ([]*OrderGetJSON, error) {
	var results []*OrderGetJSON
	rows, err := handler.db.Query(
		"SELECT order_id, accrual, status, uploadet_at FROM orders where user_id =$1", userID)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil, nil
		}
		log.Println(err.Error())
		return nil, err
	}
	for rows.Next() {
		tmp := &OrderGetJSON{}
		rows.Scan(
			&tmp.Order, &tmp.Accrual, &tmp.Status, &tmp.Processed_at)
		results = append(results, tmp)
	}

	return results, nil
}

func (handler DBHandler) RegisterOrder(orderID string, UserID string, accrual float64, status string) error {

	_, err := handler.db.Exec("INSERT INTO orders (order_id, user_id, accrual, status) VALUES($1,$2,$3,$4)", orderID, UserID, accrual, status)
	if err != nil {
		return err
	}
	return nil
}
