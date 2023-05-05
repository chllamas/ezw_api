package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	_ "github.com/joho/godotenv/autoload"
)

var secretKey []byte
var serverDB *sql.DB
var getUserStmt *sql.Stmt
var usernameExistsStmt *sql.Stmt

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")

        // Check if header present
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        // Parse JWT token
        token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return secretKey, nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(*Claims); ok && token.Valid {
            c.Set("username", claims.Username)
            c.Next()
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
        }
    }
}

func loginHandler(c *gin.Context) {
    var body struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // TODO: Change to check on database for username then check that password is correct!
    if body.Username != "bhogus" || body.Password != "dev123" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
        Username: body.Username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
        },
    })

    tokenString, err := token.SignedString(secretKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func signupHandler(c *gin.Context) {
    var err error
    var body struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err = c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // sanitize username and password inputs

    var count int
    err = usernameExistsStmt.QueryRow(body.Username).Scan(&count)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err})
        return
    } else if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
        return
    }

    // now do the hashing stuff
    // store onto servers and done!

    c.JSON(http.StatusOK, gin.H{})
}

func prepareStmts() error {
    var err error

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

    router := gin.Default()
    router.POST("/login", loginHandler)
    router.POST("/signup", signupHandler)
    router.GET("/ping", authMiddleware(), handle_pong)
    router.Run("127.0.0.1:8000")

    log.Println("End server execution")
}

