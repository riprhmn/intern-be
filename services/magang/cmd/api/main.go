package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "magang-be/services/magang/api/v1"
	"magang-be/services/magang/config"
	"magang-be/services/magang/internal/database"
	"magang-be/services/magang/internal/handler"
	"magang-be/services/magang/internal/repository"
	"magang-be/services/magang/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.Load("config/config.yaml")
	database.Init()

	userRepo := repository.NewUserRepository(database.DB)
	userSvc := service.NewUserService(userRepo)

	itemRepo := repository.NewItemRepository(database.DB)
	itemSvc := service.NewItemService(itemRepo)

	changeNoteRepo := repository.NewChangeNoteRepository(database.DB)
	changeNoteSvc := service.NewChangeNoteService(changeNoteRepo)

	authHandler := handler.NewAuthHandler(userSvc, config.App.Auth.JWTSecret)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "magang", "status": "ok"})
	})

	r.POST("/api/v1/magang/auth/login", authHandler.Login)

	v1.SetupRouter(r, itemSvc, userSvc, changeNoteSvc, authHandler, config.App.Auth.JWTSecret)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.App.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("magang service running on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down magang service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("magang service exited")
}
