package main

import (
	"BorodachevAV/gophermart/internal/config"
	"BorodachevAV/gophermart/internal/database"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
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
	database.CreateChema(handler.DBhandler)
	router := chi.NewRouter()
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.registerPost)
		r.Post("/login", handler.loginPost)
		r.Get("/orders", handler.ordersGet)
		r.Post("/orders", handler.ordersPost)
		r.Get("/balance", handler.balanceGet)
		r.Post("/balance/withdraw", handler.withdrawPost)
		r.Get("/withdrawals", handler.withdrawalsGet)
	})
	go handler.OrderProcessLoop()
	log.Fatal(http.ListenAndServe(conf.Cfg.ServerAddress, router))
}
