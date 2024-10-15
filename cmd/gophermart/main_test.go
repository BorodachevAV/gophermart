package main

import (
	"BorodachevAV/gophermart/internal/config"
	"BorodachevAV/gophermart/internal/database"
	"BorodachevAV/gophermart/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGophermart(t *testing.T) {
	conf := config.InitParams()

	handler := Handler{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := database.InitDB(conf.Cfg.DataBaseDNS, ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	handler.DBhandler = db
	handler.AccrualAddress = conf.Cfg.AccrualAddress

	t.Run("PostRegister 200", func(t *testing.T) {
		registerJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}

		reqBody, _ := json.Marshal(registerJSON)
		bodyReader := bytes.NewReader(reqBody)
		request := httptest.NewRequest(http.MethodPost, "/api/user/register", bodyReader)

		w := httptest.NewRecorder()
		handler.registerPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("PostRegister 409", func(t *testing.T) {
		registerJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}

		reqBody, _ := json.Marshal(registerJSON)
		bodyReader := bytes.NewReader(reqBody)
		request := httptest.NewRequest(http.MethodPost, "/api/user/register", bodyReader)

		w := httptest.NewRecorder()
		handler.registerPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 409, res.StatusCode)
	})

	t.Run("PostLogin 200", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)
		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)

		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("PostLogin 401", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "123",
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)
		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)

		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 401, res.StatusCode)
	})

	t.Run("PostOrders 202", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}
		orderID := "1502747701203"

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		order := strings.NewReader(orderID)
		request = httptest.NewRequest(http.MethodPost, "/api/user/orders", order)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.ordersPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 202, res.StatusCode)
	})

	t.Run("PostOrders 200", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}
		orderID := "1502747701203"

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		order := strings.NewReader(orderID)
		request = httptest.NewRequest(http.MethodPost, "/api/user/orders", order)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.ordersPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("PostOrders 401", func(t *testing.T) {

		orderID := "1502747701203"

		order := strings.NewReader(orderID)
		request := httptest.NewRequest(http.MethodPost, "/api/user/orders", order)
		w := httptest.NewRecorder()
		handler.ordersPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 401, res.StatusCode)
	})

	t.Run("PostOrders 422", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}
		orderID := "150274770103"

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		order := strings.NewReader(orderID)
		request = httptest.NewRequest(http.MethodPost, "/api/user/orders", order)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.ordersPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 422, res.StatusCode)
	})

	t.Run("GetOrders 200", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		request = httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.ordersGet(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("GetOrders 401", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		w := httptest.NewRecorder()
		handler.ordersGet(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 401, res.StatusCode)
	})

	t.Run("GetBalance 200", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		request = httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.balanceGet(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	t.Run("GetBalance 401", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		w := httptest.NewRecorder()
		handler.balanceGet(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 401, res.StatusCode)
	})

	t.Run("PostWithdraw 401", func(t *testing.T) {

		JSON := models.WithdrawalPost{
			Order: "2377225624",
			Sum:   100,
		}

		reqBody, _ := json.Marshal(JSON)

		withdrawal := bytes.NewReader(reqBody)
		request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawal)
		w := httptest.NewRecorder()
		handler.withdrawPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 401, res.StatusCode)
	})

	t.Run("PostWithdraw 402", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}
		JSON := models.WithdrawalPost{
			Order: "2377225624",
			Sum:   100,
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		reqBody, _ = json.Marshal(JSON)

		withdrawal := bytes.NewReader(reqBody)
		request = httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawal)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.withdrawPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 402, res.StatusCode)
	})

	t.Run("PostWithdraw 422", func(t *testing.T) {
		loginJSON := models.UserRequest{
			Login:    "test",
			Password: "test",
		}
		JSON := models.WithdrawalPost{
			Order: "150274770103",
			Sum:   100,
		}

		reqBody, _ := json.Marshal(loginJSON)
		bodyReader := bytes.NewReader(reqBody)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
		w := httptest.NewRecorder()
		handler.loginPost(w, request)
		cookies := w.Result().Cookies()
		reqBody, _ = json.Marshal(JSON)

		withdrawal := bytes.NewReader(reqBody)
		request = httptest.NewRequest(http.MethodPost, "/api/user//balance/withdraw", withdrawal)
		request.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		handler.ordersPost(w, request)
		res := w.Result()
		res.Body.Close()
		assert.Equal(t, 422, res.StatusCode)
	})

	t.Run("GetWithdrawals 200", func(t *testing.T) {
		t.Run("GetOrders 200", func(t *testing.T) {
			loginJSON := models.UserRequest{
				Login:    "test",
				Password: "test",
			}

			reqBody, _ := json.Marshal(loginJSON)
			bodyReader := bytes.NewReader(reqBody)

			request := httptest.NewRequest(http.MethodPost, "/api/user/login", bodyReader)
			w := httptest.NewRecorder()
			handler.loginPost(w, request)
			cookies := w.Result().Cookies()
			request = httptest.NewRequest(http.MethodGet, "/api/user/balance/Withdrawals", nil)
			request.AddCookie(cookies[0])
			w = httptest.NewRecorder()
			handler.withdrawalsGet(w, request)
			res := w.Result()
			res.Body.Close()
			assert.Equal(t, 200, res.StatusCode)
		})
	})
}
