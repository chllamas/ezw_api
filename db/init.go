package db

import (
	"errors"
	"regexp"

    "gorm.io/gorm"
	"gorm.io/driver/mysql"
)

var secretKey []byte
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
    Hash        []byte    `gorm:"type:BLOB(32);not null"`
    Salt        []byte    `gorm:"type:BLOB(32);not null"`
}

func TODO() error {
    return errors.New("TODO: function not yet implemented")
}

func Init(dsn string, sk string) {
    if database != nil {
        panic("Datbase init can only be called once!")
    }
    var err error
    secretKey = []byte(sk)
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

func GetSecretKey() []byte {
    return secretKey
}

func ValidateUsername(u string) bool {
    return usernameSanitizer.MatchString(u)
}

func ValidatePassword(p string) bool {
    return passwordSanitizer.MatchString(p)
}

func CreateUser(u *User) error {
    if result := database.Create(u); result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return errors.New("username already exists")
        }
        return result.Error
    }
    return nil
}

func ReadUser(username string, body *User) error {
    if err := database.Take(body, "username = ?", username).Error; err != nil {
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
