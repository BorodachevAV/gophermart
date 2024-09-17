package main

import (
	"BorodachevAV/gophermart/internal/auth"
	"BorodachevAV/gophermart/internal/database"
	"BorodachevAV/gophermart/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gospacedev/luhn"
)

type Handler struct {
	DBhandler      *database.DBHandler
	AccrualAddress string
}

func (handler Handler) getAccrual(orderID string) (*models.AccrualJSONRequest, error) {
	var req *models.AccrualJSONRequest
	var buf bytes.Buffer

	requestURL := fmt.Sprintf("%s/api/orders/%s", handler.AccrualAddress, orderID)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Println("get accrual error", err.Error())
		return nil, err
	}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		log.Println("accrual read response body error", err.Error())
		return nil, err
	}
	defer res.Body.Close()
	if len(buf.Bytes()) == 0 {
		log.Println("accrual response is empty")
		return nil, nil
	}

	log.Println("read json")
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		log.Println("accrual read json error", err.Error())
		return nil, err
	}
	return req, nil
}

func (handler Handler) registerPost(w http.ResponseWriter, r *http.Request) {
	var req models.UserJSONRequest
	var buf bytes.Buffer
	log.Println("read body")
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("read json")
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("duplicate", req.Login)
	isDuplicate, err := handler.DBhandler.CheckDuplicateLogin(req.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isDuplicate {
		http.Error(w, "login is taken", http.StatusConflict)
		return
	}

	cookie, err := auth.BuildJWTString()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sha256Pass, err := auth.SHA256password(req.Password)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("register")
	err = handler.DBhandler.RegisterUser(req.Login, sha256Pass, cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cookies := &http.Cookie{
		Name:  "UserID",
		Value: cookie,
	}
	http.SetCookie(w, cookies)
	w.WriteHeader(http.StatusOK)
}

func (handler Handler) loginPost(w http.ResponseWriter, r *http.Request) {
	var req models.UserJSONRequest
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sha256Pass, err := auth.SHA256password(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pass, err := handler.DBhandler.GetUserPassword(req.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pass == "" {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if pass != sha256Pass {
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}
	userID, err := handler.DBhandler.GetUserID(req.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cookies := &http.Cookie{
		Name:  "UserID",
		Value: userID,
	}
	http.SetCookie(w, cookies)
	w.WriteHeader(http.StatusOK)
}

func (handler Handler) ordersGet(w http.ResponseWriter, r *http.Request) {
	log.Println("ordersGet called")
	userID, err := r.Cookie("UserID")

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	orders, err := handler.DBhandler.GetOrdersByUserID(userID.Value)
	log.Println(orders[0])
	respBody, _ := json.Marshal(orders)

	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err.Error())
	}

}

func (handler Handler) ordersPost(w http.ResponseWriter, r *http.Request) {

	userID, err := r.Cookie("UserID")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("order number read error", err.Error())
	}
	orderID := string(body)
	if _, err := strconv.ParseInt(orderID, 10, 64); err != nil {
		http.Error(w, "not numeric", http.StatusUnprocessableEntity)
		return
	}

	if !luhn.Check(orderID) {
		http.Error(w, "not numeric", http.StatusUnprocessableEntity)
		return
	}

	//check if used
	log.Println("check if used")
	DBUserID, err := handler.DBhandler.GetUseIDByOrderID(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if DBUserID != "" {
		if DBUserID == userID.Value {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			log.Println("order already registered", orderID)
			http.Error(w, "order already registered", http.StatusConflict)
			return
		}
	} else {
		accrual, err := handler.getAccrual(orderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("register order", orderID)
		var acc float64
		var status string
		if accrual == nil {
			log.Println("accrual empty response", orderID)
			acc = 0
			status = "NEW"
		} else {
			acc = accrual.Accrual
			status = accrual.Status
		}

		err = handler.DBhandler.RegisterOrder(orderID, userID.Value, acc, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		balance, err := handler.DBhandler.GetBalance(userID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		balance = balance + accrual.Accrual
		err = handler.DBhandler.SetBalance(userID.Value, balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}

}

func (handler Handler) balanceGet(w http.ResponseWriter, r *http.Request) {
	userID, err := r.Cookie("UserID")

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	balance, err := handler.DBhandler.GetBalance(userID.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	withdrawals_sum, err := handler.DBhandler.GetWithdrawalsSum(userID.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := models.BalanceGetJSON{
		Current:   balance,
		Withdrawn: withdrawals_sum,
	}
	respBody, _ := json.Marshal(response)

	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err.Error())
	}

}

func (handler Handler) withdrawPost(w http.ResponseWriter, r *http.Request) {
	var req models.WIthdrawJSONRequest
	var buf bytes.Buffer

	userID, err := r.Cookie("UserID")

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("read json")
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseInt(req.Order, 10, 64); err != nil {
		http.Error(w, "not numeric", http.StatusUnprocessableEntity)
		return
	}

	if !luhn.Check(req.Order) {
		http.Error(w, "not numeric", http.StatusUnprocessableEntity)
		return
	}

	balance, err := handler.DBhandler.GetBalance(userID.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Sum > balance {
		http.Error(w, "balance too low", http.StatusPaymentRequired)
	}

	balance = balance - req.Sum

	err = handler.DBhandler.SetBalance(userID.Value, balance)
	if err != nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = handler.DBhandler.RegisterWithdrawal(req.Order, userID.Value, req.Sum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (handler Handler) withdrawalsGet(w http.ResponseWriter, r *http.Request) {

	userID, err := r.Cookie("UserID")

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	response, err := handler.DBhandler.GetUserWithdrawals(userID.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respBody, _ := json.Marshal(response)

	_, err = w.Write(respBody)
	if err != nil {
		log.Println(err.Error())
	}

}
