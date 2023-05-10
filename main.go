package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/chllamas/ezw_api/auth"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

var serverDB *sql.DB
var getUserStmt *sql.Stmt
var insertUserStmt *sql.Stmt
var usernameExistsStmt *sql.Stmt
var packageStmts map[string]*sql.Stmt

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func prepareStmts() error {
    var err error

    insertUserStmt, err = serverDB.Prepare("INSERT INTO users (username, password, salt) VALUES (?, ?, ?)")
    if err != nil {
        return err
    }

    getUserStmt, err = serverDB.Prepare("SELECT user_id, username, password, salt FROM users WHERE username = ?")
    if err != nil {
        return err
    }

    usernameExistsStmt, err = serverDB.Prepare("SELECT COUNT(*) FROM users WHERE username = ?")
    if err != nil {
        return err
    }

    return nil
}

func main() {
    var err error
    var secretKey []byte

    if skStr, ok := os.LookupEnv("JWT_SECRET_KEY"); ok {
        secretKey = []byte(skStr)
    } else {
        log.Fatalf("Secret key is not set")
    }

    if serverDB, err = sql.Open("mysql", os.Getenv("DSN")); err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer serverDB.Close()

    if err = prepareStmts(); err != nil {
        log.Fatalf("failed to prepare stmts: %v", err)
    }
    defer getUserStmt.Close()
    defer insertUserStmt.Close()
    defer usernameExistsStmt.Close()

    packageStmts = map[string]*sql.Stmt{
        "getUserStmt":getUserStmt,
        "insertUserStmt":insertUserStmt,
        "usernameExistsStmt":usernameExistsStmt,
    }

    router := gin.Default()
    router.POST("/login", auth.LoginHandler(secretKey, &packageStmts))
    router.POST("/signup", auth.SignupHandler(secretKey, &packageStmts))
    router.GET("/ping", auth.AuthMiddleware(secretKey), handle_pong)
    router.Run("0.0.0.0:8000")

    log.Println("End server execution")
}

