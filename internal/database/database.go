package database

import (
	"BorodachevAV/gophermart/internal/models"
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
			accrual  float,
			status VARCHAR(200),
			uploadet_at timestamp default current_timestamp
		)`

	createTableBalance :=
		`CREATE TABLE IF NOT EXISTS balance(
			user_id VARCHAR(200) PRIMARY KEY,
			balance  float
		)`

	createTableBalanceLog :=
		`CREATE TABLE IF NOT EXISTS Withdrawals_log(
			order_id VARCHAR(200),
			user_id VARCHAR(200) NOT NULL,
			withdrawal float,
			processed_at timestamp default current_timestamp
		)`

	_, err := db.db.Exec(createTableUsers)
	if err != nil {
		return err
	}
	_, err = db.db.Exec(createTableOrders)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(createTableBalance)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(createTableBalanceLog)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) SetStatus(userID string, status string) error {

	_, err := handler.db.Exec("UPDATE orders set status = $1 where user_id = $2", status, userID)
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

func (handler DBHandler) GetOrdersByUserID(userID string) ([]models.OrderGet, error) {
	var results []models.OrderGet
	rows, err := handler.db.Query(
		"SELECT order_id, accrual, status, uploadet_at FROM orders where user_id =$1", userID)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil, nil
		}
		log.Println(err.Error())
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		tmp := models.OrderGet{}
		rows.Scan(
			&tmp.Order, &tmp.Accrual, &tmp.Status, &tmp.ProcessedAt)
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

func (handler DBHandler) SetOrderAccrual(orderID string, accrual float64) error {
	_, err := handler.db.Exec("Update orders set accrual = ($1) where order_id = $2", accrual, orderID)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) GetBalance(userID string) (float64, error) {
	var balance float64

	err := handler.db.QueryRow(
		"SELECT balance FROM balance where user_id =$1", userID).Scan(&balance)
	if err != nil {

		if err.Error() == sql.ErrNoRows.Error() {

			return 0, nil
		}
		log.Println("GetBalance error")
		return 0, err
	}
	return balance, nil
}

func (handler DBHandler) SetBalance(userID string, balance float64) error {
	var currentBalance float64
	err := handler.db.QueryRow(
		"SELECT balance FROM balance where user_id =$1", userID).Scan(&currentBalance)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			_, err := handler.db.Exec("Insert into balance (balance, user_id) values ($1, $2)", balance, userID)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	_, err = handler.db.Exec("UPDATE balance set balance = $1 where user_id = $2", balance, userID)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) GetWithdrawalsSum(userID string) (float64, error) {
	var withdrawalSum float64
	var HasUser string
	err := handler.db.QueryRow(
		"SELECT user_id FROM withdrawals_log where user_id =$1", userID).Scan(&HasUser)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return 0, nil
		}
		log.Println(err.Error())
		return 0, err
	}

	err = handler.db.QueryRow(
		"SELECT SUM(withdrawal) FROM withdrawals_log where user_id =$1", userID).Scan(&withdrawalSum)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	return withdrawalSum, nil
}

func (handler DBHandler) RegisterWithdrawal(orderID string, userID string, withdrawal float64) error {

	_, err := handler.db.Exec("INSERT INTO Withdrawals_log (order_id, user_id, withdrawal) VALUES($1,$2,$3)", orderID, userID, withdrawal)
	if err != nil {
		return err
	}
	return nil
}

func (handler DBHandler) GetUserWithdrawals(userID string) ([]models.WithdrawalGet, error) {
	var results []models.WithdrawalGet
	rows, err := handler.db.Query("SELECT order_id, withdrawal, processed_at FROM Withdrawals_log where user_id =$1", userID)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return results, nil
		}
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		tmp := models.WithdrawalGet{}
		err = rows.Scan(
			&tmp.Order, &tmp.Withdrawal, &tmp.ProcessedAt)
		if err != nil {
			return nil, err
		}
		log.Println(tmp)
		results = append(results, tmp)

	}
	return results, nil

}
