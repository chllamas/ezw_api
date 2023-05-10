package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var usernameSanitizer = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._]{1,30}[a-zA-Z0-9])?$`)
var passwordSanitizer = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*?]{8,128}$`)

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

type UserLogin struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func AuthMiddleware(secretKey []byte) gin.HandlerFunc {
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

func LoginHandler(secretKey []byte, stmts *map[string]*sql.Stmt) gin.HandlerFunc {
    return func(c *gin.Context) {
        var err error
        var body UserLogin
        var storedData struct {
            User_ID  int    `json:"userid"`
            Username string `json:"username"`
            Password []byte `json:"password"`
            Salt     []byte `json:"salt"`
        }

        if err = c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if !usernameSanitizer.MatchString(body.Username) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "username is not valid"})
            return
        } else if !passwordSanitizer.MatchString(body.Password) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "password is not valid"})
            return
        }

        err = (*stmts)["getUserStmt"].QueryRow(body.Username).Scan(&storedData.User_ID, &storedData.Username, &storedData.Password, &storedData.Salt)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
            return
        }

        salt := storedData.Salt
        passwd := []byte(body.Password)
        passwd = append(passwd, salt...)
        hash := sha256.Sum256(passwd)
        hashedPasswd := hash[:]

        if !reflect.DeepEqual(hashedPasswd, storedData.Password) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
            return
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
            Username: body.Username,
            StandardClaims: jwt.StandardClaims{
                ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
            },
        })

        var tokenString string
        tokenString, err = token.SignedString(secretKey)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"token": tokenString})
    }
}

func SignupHandler(secretKey []byte, stmts *map[string]*sql.Stmt) gin.HandlerFunc {
    return func(c *gin.Context) {
        var err error
        var body UserLogin

        if err = c.ShouldBindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        username_param_str := "Username: 3-32 chars, alphanumerics & special chars: _."
        password_param_str := "Password: 8-128 chars, alphanumerics & special chars: !@#$%^&*?"

        if !usernameSanitizer.MatchString(body.Username) {
            c.JSON(http.StatusBadRequest, gin.H{"error": username_param_str})
            return
        } else if !passwordSanitizer.MatchString(body.Password) {
            c.JSON(http.StatusBadRequest, gin.H{"error": password_param_str})
            return
        }

        var count int
        err = (*stmts)["usernameExistsStmt"].QueryRow(body.Username).Scan(&count)
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

        _, err = (*stmts)["insertUserStmt"].Exec(body.Username[:], hashedPasswd, salt)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        } else {
            c.JSON(http.StatusOK, gin.H{})
        }
    }
}
