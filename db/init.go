package db

import (
    "regexp"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"

    . "github.com/chllamas/go-utils"
)

var db *gorm.DB
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

func Init(secretKey string) error {
    var err error
    db, err = gorm.Open(mysql.New(mysql.Config{
        DSN: secretKey,
        DefaultStringSize: 256,
    }), &gorm.Config{})

    return err
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

func Close() {
    dbInstance, _ := db.DB()
    dbInstance.Close()
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
