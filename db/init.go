package db

import (
	"errors"
	"regexp"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var database *gorm.DB
var secretKey string
var usernameSanitizer = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._]{1,30}[a-zA-Z0-9])?$`)
var passwordSanitizer = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*?]{8,128}$`)

const UsernameParameterString = "Username: 3-32 chars, alphanumerics & special chars: _."
const PasswordParameterString = "Password: 8-128 chars, alphanumerics & special chars: !@#$%^&*?"

type UserLogin struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type User struct {
    Username    string     `gorm:"primaryKey"`
    Hash        [32]byte
    Salt        [32]byte
}

func TODO() error {
    return errors.New("TODO: function not yet implemented")
}

func Init(dsn string, sk string) {
    var err error
    secretKey = sk
    database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }
}

func Close() {
    dbInstance, _ := database.DB()
    dbInstance.Close()
}

func GetSecretKey() string {
    return secretKey
}

func ValidateUsername(u string) bool {
    return usernameSanitizer.MatchString(u)
}

func ValidatePassword(p string) bool {
    return passwordSanitizer.MatchString(p)
}

func UsernameExists(username string) error {
    return TODO()
}

func CreateUser(username string, hash []byte, salt []byte) error {
    return TODO()
}

func ReadUser(username string, body *User) error {
    return TODO()
}

func UpdateUser(username string, new_hash []byte, new_salt []byte) error {
    return TODO()
}

func DeleteUser(username string) error {
    return TODO()
}
