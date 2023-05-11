package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"net/http"
	"reflect"
	"time"

	"github.com/chllamas/ezw_api/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

type HashTuple struct {
    Hash [32]byte
    Salt [32]byte
}

func AuthMiddleware() gin.HandlerFunc {
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
            return db.GetSecretKey(), nil
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

func HashPassword(p string, s *[32]byte) (*HashTuple, error) {
    var salt [32]byte
    if s == nil {
        if _, err := rand.Read(salt[:]); err != nil {
            return nil, err
        }
    } else {
        salt = *s
    }

    passwd := append([]byte(p), salt[:]...)
    hash := sha256.Sum256(passwd)

    ret := HashTuple{
        Hash: hash,
        Salt: salt,
    }

    return &ret, nil
}

func LoginHandler(c *gin.Context) {
    var body db.UserLogin
    var userData db.User

    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if !db.ValidateUsername(body.Username) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "input not valid"})
        return
    } else if !db.ValidatePassword(body.Password) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "input not valid"})
        return
    }

    if err := db.ReadUser(body.Username, &userData); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    if hash, err := HashPassword(body.Password, &userData.Salt); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    } else if !reflect.DeepEqual(hash.Hash, userData.Hash) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
        Username: body.Username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
        },
    })

    if tokenString, err := token.SignedString(db.GetSecretKey()); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
    } else {
        c.JSON(http.StatusOK, gin.H{"token": tokenString})
    }
}

func SignupHandler(c *gin.Context) {
    var body db.UserLogin

    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if !db.ValidateUsername(body.Username) {
        c.JSON(http.StatusBadRequest, gin.H{"error": db.UsernameParameterString})
        return
    } else if !db.ValidatePassword(body.Password) {
        c.JSON(http.StatusBadRequest, gin.H{"error": db.PasswordParameterString})
        return
    }

    if err := db.UsernameExists(body.Username); err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        return
    }

    salt := make([]byte, 32) 
    if _, err := rand.Read(salt); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "password creation failed"})
        return
    }

    passwd := []byte(body.Password)
    passwd = append(passwd, salt...)
    hash := sha256.Sum256(passwd)
    hashedPasswd := hash[:]


    if err := db.CreateUser(body.Username, hashedPasswd, salt); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // TODO: Maybe return a token for them since they are now logged in when creating account?
    c.JSON(http.StatusOK, gin.H{})
}
