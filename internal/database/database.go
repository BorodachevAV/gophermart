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
			accrual  int,
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

func (handler DBHandler) RegisterOrder(orderID string, UserID string, accrual int) error {

	_, err := handler.db.Exec("INSERT INTO orders (order_id, user_id, accrual) VALUES($1,$2,$3)", orderID, UserID, accrual)
	if err != nil {
		return err
	}
	return nil
}
