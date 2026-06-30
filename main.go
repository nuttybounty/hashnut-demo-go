package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"hashnut-demo-shop/internal/config"
	"hashnut-demo-shop/internal/handler"
	"hashnut-demo-shop/internal/notify"
	"hashnut-demo-shop/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	s := store.New(db)
	h := handler.New(s, &cfg.HashNut)
	n := notify.New(s)

	r := gin.Default()
	r.Use(cors.Default())

	// API routes
	api := r.Group("/api")
	{
		api.GET("/products", h.ListProducts)
		api.GET("/chains", h.ListChains)
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/:id", h.GetOrder)
		api.POST("/orders/:id/confirm", h.ConfirmPaid)
		api.POST("/notify", n.HandleNotify)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Demo shop starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
