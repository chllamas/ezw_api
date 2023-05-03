package main

import (
    "os"
    "log"
    "net/http"
    "database/sql"
    "github.com/gin-gonic/gin"
     _ "github.com/go-sql-driver/mysql"
     _ "github.com/joho/godotenv/autoload"
)

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func main() {
    db, err := sql.Open("mysql", os.Getenv("DSN"))
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("failed to ping: %v", err)
    }

    router := gin.Default()
    router.GET("/ping", handle_pong)
    router.Run("127.0.0.1:8000")
}

