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

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func prepareStmts(db *sql.DB) (map[string]*sql.Stmt, error) {
    var err error
    ret := make(map[string]*sql.Stmt, 3)

    ret["insertUserStmt"], err = db.Prepare("INSERT INTO users (username, password, salt) VALUES (?, ?, ?)")
    if err != nil {
        return nil, err
    }

    ret["getUserStmt"], err = db.Prepare("SELECT user_id, username, password, salt FROM users WHERE username = ?")
    if err != nil {
        return nil, err
    }

    ret["usernameExistsStmt"], err = db.Prepare("SELECT COUNT(*) FROM users WHERE username = ?")
    if err != nil {
        return nil, err
    }

    return ret, nil
}

func main() {
    var err error
    var db *sql.DB
    var secretKey []byte
    var stmts map[string]*sql.Stmt

    if skStr, ok := os.LookupEnv("JWT_SECRET_KEY"); ok {
        secretKey = []byte(skStr)
    } else {
        log.Fatalf("Secret key is not set")
    }

    if db, err = sql.Open("mysql", os.Getenv("DSN")); err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer db.Close()

    if stmts, err = prepareStmts(db); err != nil {
        log.Fatalf("failed to prepare stmts: %v", err)
    }
    defer func(){
        for _,v := range stmts {
            v.Close()
        }
    }()

    router := gin.Default()
    router.POST("/login", auth.LoginHandler(secretKey, &stmts))
    router.POST("/signup", auth.SignupHandler(secretKey, &stmts))
    router.GET("/ping", auth.AuthMiddleware(secretKey), handle_pong)
    router.Run("0.0.0.0:8000")

    log.Println("End server execution")
}

