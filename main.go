package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/a-h/templ/examples/integration-gin/gintemplrenderer"
	"github.com/chllamas/ezw_api/auth"
	"github.com/chllamas/ezw_api/db"
	"github.com/chllamas/ezw_api/templates"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)


/* Returns username, id values if they exist, nil otherwise; it's up to caller to handle if a missing value results in an error */
func getUserInfo(c *gin.Context) (*string, *int) {
    var username *string
    var id *int

    if usernameRaw, exists := c.Get("username"); exists {
        if u, ok := usernameRaw.(string); ok {
            username = &u
        }
    }

    if idStr := c.Param("id"); idStr != "" {
        if i, err := strconv.Atoi(idStr); err == nil {
            id = &i
        }
    }

    return username, id
}

func handleNewTasks(c *gin.Context) {
    username, _ := getUserInfo(c)
    if username == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
        return
    }

    var body db.CreateTaskRequest
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := db.CreateTask(*username, &body); err != nil {
        c.JSON(err.HttpStatus(), gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{})
}

func handleGetTasks(c *gin.Context) {
    username, id := getUserInfo(c)
    if username == nil || id == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
        return
    }

    ret, err := db.ReadTask(*username, *id)
    if err != nil {
        c.JSON(err.HttpStatus(), gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"results": ret})
}

func handleGetAllTasks(c *gin.Context) {
    username, _ := getUserInfo(c)
    if username == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
        return
    }

    ret, err := db.ReadAllTasks(*username)
    if err != nil {
        c.JSON(err.HttpStatus(), gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"results": ret})
}

func handleEditTasks(c *gin.Context) {
    username, id := getUserInfo(c)
    if username == nil || id == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
        return
    }

    var body db.UpdateTaskRequest
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := db.UpdateTask(*username, *id, &body); err != nil {
        c.JSON(err.HttpStatus(), gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{})
}

func handleDeleteTasks(c *gin.Context) {
    username, id := getUserInfo(c)
    if username == nil || id == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user info"})
        return
    }

    if err := db.DeleteTask(*username, *id); err != nil {
        c.JSON(err.HttpStatus(), gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{})
}

func handleLandingPage(c *gin.Context) {
    c.HTML(http.StatusOK, "", templates.LandingPage("dark", "EZW - The Task Manager"))
}

func handleLoginPage(c *gin.Context) {
    c.HTML(http.StatusOK, "", templates.TempPage("login"))
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
    router.HTMLRender = gintemplrenderer.Default

    router.GET("/", handleLandingPage)
    router.GET("/login", handleLoginPage)

    router.POST("/api/login", auth.LoginHandler)
    router.POST("/api/signup", auth.SignupHandler)

    router.POST("/api/tasks", auth.AuthMiddleware(), handleNewTasks)
    router.GET("/api/tasks", auth.AuthMiddleware(), handleGetAllTasks)
    router.GET("/api/tasks/:id", auth.AuthMiddleware(), handleGetTasks)
    router.PUT("/api/tasks/:id", auth.AuthMiddleware(), handleEditTasks)
    router.DELETE("/api/tasks/:id", auth.AuthMiddleware(), handleDeleteTasks)

    router.GET("/api/ping", auth.AuthMiddleware(), func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "pong"})
    })
    router.Run(":8080")
}

