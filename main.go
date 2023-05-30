package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
    "errors"

	"github.com/chllamas/ezw_api/auth"
	"github.com/chllamas/ezw_api/db"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func extractUsername(c *gin.Context) (string, error) {
    usernameRaw, exists := c.Get("username")
    if !exists {
        return "", errors.New("Couldn't find username in JSON")
    }
    
    username, ok := usernameRaw.(string)
    if !ok {
        return "", errors.New("failed to parse username to string")
    }

    return username, nil
}

func handleGetTasks(c *gin.Context) {
    username, err := extractUsername(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
    username, err := extractUsername(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var body db.TaskRequest
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := db.CreateTask(username, &body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{})
}

func handleEditTasks(c *gin.Context) {
    idStr := c.Param("id")

    if id, err := strconv.Atoi(idStr); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    } else {
        username, err := extractUsername(c)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var body struct {
            NewTitle            string  `json:"new_title"`
            NewContents         string  `json:"new_contents"`
            ToggleCompleted     bool    `json:"toggle_completed"`
        }

        if err := c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        var request db.TaskRequest
        task, err := db.ReadTask(username, id)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }

        if body.NewTitle == "" {
            request.Title = task.Title
        } else {
            request.Title = body.NewTitle
        }

        if body.NewContents == "" {
            request.Contents = task.Contents
        } else {
            request.Contents = body.NewContents
        }

        if body.ToggleCompleted {
            request.Completed = !task.Completed
        }
            
        if err := db.UpdateTask(username, id, &request); err != nil {
            c.JSON(http.StatusNotFound, gin.H{"err": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{})
    }
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
    router.PUT("/tasks/:id", auth.AuthMiddleware(), handleEditTasks)
    router.GET("/tasks/:id", auth.AuthMiddleware(), handleGetTasks)
    router.GET("/tasks", auth.AuthMiddleware(), handleGetTasks)
    router.POST("/tasks", auth.AuthMiddleware(), handleNewTasks)
    router.Run("0.0.0.0:8000")
}

