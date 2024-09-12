package main

import (
	"BorodachevAV/gophermart/internal/auth"
	"BorodachevAV/gophermart/internal/database"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	DBhandler      *database.DBHandler
	AccrualAddress string
}

type UserJSONRequest struct {
	Login    string
	Password string
}

type AccrualJSONRequest struct {
	Order   string
	Status  string
	Accrual int
}

func (handler Handler) getAccrual(orderID string) (*AccrualJSONRequest, error) {
	var req *AccrualJSONRequest
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
	var req UserJSONRequest
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
	var req UserJSONRequest
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
	if _, err := strconv.Atoi(orderID); err != nil {
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
			http.Error(w, "order already registered", http.StatusConflict)
			return
		}
	} else {
		accrual, err := handler.getAccrual(orderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if accrual == nil {
			log.Println("accrual empty response")
			http.Error(w, "accrual empty response", http.StatusInternalServerError)
			return
		}
		log.Println("register order")
		err = handler.DBhandler.RegisterOrder(orderID, userID.Value, accrual.Accrual)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}

}

func (handler Handler) balanceGet(w http.ResponseWriter, r *http.Request) {

}

func (handler Handler) withdrawPost(w http.ResponseWriter, r *http.Request) {

}

func (handler Handler) withdrawalsGet(w http.ResponseWriter, r *http.Request) {

}
