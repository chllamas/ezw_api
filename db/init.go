package db

import (
	"errors"
	"regexp"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var secretKey string
var database *gorm.DB = nil
var usernameSanitizer = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._]{1,30}[a-zA-Z0-9])?$`)
var passwordSanitizer = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*?]{8,128}$`)

const UsernameParameterString = "Username: 3-32 chars, alphanumerics & special chars: _."
const PasswordParameterString = "Password: 8-128 chars, alphanumerics & special chars: !@#$%^&*?"

type UserLogin struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type User struct {
    Username    string      `gorm:"type:varchar(32);primaryKey"`
    Hash        [32]byte    `gorm:"type:binary(32);not null"`
    Salt        [32]byte    `gorm:"type:binary(32);not null"`
}

func TODO() error {
    return errors.New("TODO: function not yet implemented")
}

func Init(dsn string, sk string) {
    if database != nil {
        panic("Datbase init can only be called once!")
    }
    var err error
    secretKey = sk
    database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }
    database.AutoMigrate(&User{})
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
    if err := database.First(body, "username = ?", username).Error; err != nil {
        return err
    }
    return nil
}

func UpdateUser(username string, new_hash []byte, new_salt []byte) error {
    return TODO()
}

func DeleteUser(username string) error {
    return TODO()
}
