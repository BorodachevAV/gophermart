package main

import (
	"BorodachevAV/gophermart/internal/config"
	"BorodachevAV/gophermart/internal/database"
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	conf := config.NewConfig()

	a := flag.String("a", "localhost:8080", "shortener host")
	r := flag.String("b", "http://localhost:8080", "accrual system address")
	d := flag.String("d", "postgresql://postgres:password@localhost", "db connect string")

	flag.Parse()

	if conf.Cfg.ServerAddress == "" {
		conf.Cfg.ServerAddress = *a
	}
	if conf.Cfg.AccrualAddress == "" {
		conf.Cfg.AccrualAddress = *r
	}

	if conf.Cfg.DataBaseDNS == "" {
		conf.Cfg.DataBaseDNS = *d
	}

	handler := Handler{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db, err := database.InitDB(conf.Cfg.DataBaseDNS, ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	handler.DBhandler = db
	handler.AccrualAddress = conf.Cfg.AccrualAddress
	database.CreateChema(handler.DBhandler)
	router := chi.NewRouter()
	router.Post(`/api/user/register`, handler.register_post)
	router.Post(`/api/user/login`, handler.login_post)
	router.Get(`/api/user/orders`, handler.orders_get)
	router.Post(`/api/user/orders`, handler.orders_post)
	router.Get(`/api/user/balance`, handler.balance_get)
	router.Post(`/api/user/balance/withdraw`, handler.withdraw_post)
	router.Get(`/api/user/withdrawals`, handler.withdrawals_get)

	log.Fatal(http.ListenAndServe(conf.Cfg.ServerAddress, router))
}
