package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	_ "github.com/joho/godotenv/autoload"
)

var secretKey []byte
var serverDB *sql.DB
var getUserStmt *sql.Stmt
var insertUserStmt *sql.Stmt
var usernameExistsStmt *sql.Stmt
var usernameSanitizer = regexp.MustCompile(`&[a-zA-Z0-9_.]{3,32}$`)
var passwordSanitizer = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*?]{8,128}$`)

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

        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

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

    // TODO: check if user exists, then grab that user's data 
    // TODO: hash the given password and match it to the password that's stored on DB
    // TODO: then remove the placeholder implementation below
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

    username_param_str := "Username: 3-32 chars, alphanumerics & special chars: _."
    password_param_str := "Password: 8-128 chars, alphanumerics & special chars: !@#$%^&*?"

    if err = c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if !usernameSanitizer.MatchString(body.Username) {
        c.JSON(http.StatusBadRequest, gin.H{"error": username_param_str})
        return
    } else if !passwordSanitizer.MatchString(body.Password) {
        c.JSON(http.StatusBadRequest, gin.H{"error": password_param_str})
        return
    }

    var count int
    err = usernameExistsStmt.QueryRow(body.Username).Scan(&count)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err})
        return
    } else if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
        return
    }

    salt := make([]byte, 32) 
    _, err = rand.Read(salt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "password creation failed"})
        return
    }

    passwd := []byte(body.Password)
    passwd = append(passwd, salt...)
    hash := sha256.Sum256(passwd)
    hashedPasswd := hash[:]

    insertUserStmt.Exec(body.Username[:], hashedPasswd, salt)

    c.JSON(http.StatusOK, gin.H{})
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

    router := gin.Default()
    router.POST("/login", loginHandler)
    router.POST("/signup", signupHandler)
    router.GET("/ping", authMiddleware(), handle_pong)
    router.Run("0.0.0.0:8000")

    log.Println("End server execution")
}

