package main

import (
	"log"
	"net/http"
	"os"

    "github.com/chllamas/ezw_api/db"
	"github.com/chllamas/ezw_api/auth"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func main() {
    secretKey, ok := os.LookupEnv("JWT_SECRET_KEY")
    if !ok {
        log.Fatalf("Secret key is not set")
    }

    if err := db.Init(secretKey); err != nil {
        log.Fatalf("%v", err.Error())
    }
    defer db.Close()

    router := gin.Default()
    router.POST("/login", auth.LoginHandler)
    router.POST("/signup", auth.SignupHandler)
    router.GET("/ping", auth.AuthMiddleware(), handle_pong)
    router.Run("0.0.0.0:8000")
}

