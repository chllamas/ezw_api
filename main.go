package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/chllamas/ezw_api/auth"
	"github.com/chllamas/ezw_api/db"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func handleGetTasks(c *gin.Context) {
    usernameRaw, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set username in auth middleware"})
        return
    }
    
    username, ok := usernameRaw.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse username to string"})
        return
    }

    if idStr := c.Param("id"); idStr != "" {
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
            return
        }

        ret, err := db.ReadTask(username, id)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "could not find task with given id"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"results": *ret})
        return
    }

    ret, err := db.ReadAllTasks(username)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"results": ret})
}

func handleNewTasks(c *gin.Context) {
    var body db.TaskRequest

    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    usernameRaw, exists := c.Get("username")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set username in auth middleware"})
        return
    }

    username, ok := usernameRaw.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse username to string"})
        return
    }

    if err := db.CreateTask(username, &body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{})
}

func main() {
    dsn, ok := os.LookupEnv("DSN")
    if !ok {
        log.Fatalf("DSN not set")
    }

    secretKey, ok := os.LookupEnv("JWT_SECRET_KEY")
    if !ok {
        log.Fatalf("JWT secret key is not set")
    }

    db.Init(dsn, secretKey)
    defer db.Close()

    router := gin.Default()
    router.GET("/ping", auth.AuthMiddleware(), func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "pong"})
    })
    router.POST("/login", auth.LoginHandler)
    router.POST("/signup", auth.SignupHandler)
    router.GET("/tasks/:id", auth.AuthMiddleware(), handleGetTasks)
    router.GET("/tasks", auth.AuthMiddleware(), handleGetTasks)
    router.POST("/tasks", auth.AuthMiddleware(), handleNewTasks)
    router.Run("0.0.0.0:8000")
}

