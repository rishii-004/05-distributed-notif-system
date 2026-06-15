package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/i-katta/notif-system/internal/handler"
	"github.com/i-katta/notif-system/internal/worker"
	"github.com/i-katta/notif-system/pkg/config"
)

func main() {
	cfg := config.Load()

	worker.Start(3)

	r := gin.Default()
	r.POST("/api/v1/events", handler.PostEvent)

	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
