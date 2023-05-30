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

type TaskRequest struct {
    Title       string  `json:"title" binding:"required"`
    Completed   bool    `json:"completed"`
    Contents    string  `json:"contents"`
}

type TaskHalfResponse struct {
    ID          int
    Title       string  `json:"title"`
    Completed   bool    `json:"completed"`
}

type TaskFullResponse struct {
    ID          int
    Title       string  `json:"title"`
    Contents    string  `json:"contents"`
    Completed   bool    `json:"completed"`
}

type UserLogin struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type User struct {
    Username    string      `gorm:"type:varchar(32);primaryKey"`
    Hash        []byte      `gorm:"type:BLOB(32);not null"`
    Salt        []byte      `gorm:"type:BLOB(32);not null"`
}

type Task struct {
    gorm.Model
    UserID      string      `gorm:"type:varchar(32);not null"`
    Title       string      `gorm:"type:varchar(128);not null"`
    Completed   bool        `gorm:"type:bool;not null"`
    Contents    string      `gorm:"type:longtext;not null"`
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
    database.AutoMigrate(&User{}, &Task{})
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
        return result.Error
    }
    return nil
}

func ReadUser(username string, body *User) error {
    if result := database.Take(body, "username = ?", username); result.Error != nil {
        return result.Error
    }
    return nil
}

func UpdateUser(username string, new_hash []byte, new_salt []byte) error {
    return TODO()
}

func DeleteUser(username string) error {
    return TODO()
}

func CreateTask(username string, t *TaskRequest) error {
    // create a task obj
    task := Task {
        UserID: username,
        Title: t.Title,
        Completed: t.Completed,
        Contents: t.Contents,
    }
    if result := database.Create(&task); result.Error != nil {
        return result.Error
    }
    return nil
}

func ReadTask(username string, id int) (*TaskFullResponse, error) {
    var task TaskFullResponse
    if result := database.Table("tasks").Select("id, title, completed, contents").Where("user_id = ? AND id = ?", username, id).First(&task); result.Error != nil {
        return nil, result.Error
    }
    return &task, nil
}

func ReadAllTasks(username string) ([]TaskHalfResponse, error) {
    var tasks []TaskHalfResponse
    if result := database.Table("tasks").Select("id, title, completed").Where("user_id = ?", username).Find(&tasks); result.Error != nil {
        return nil, result.Error
    }
    return tasks, nil
}

func UpdateTask(username string, id int, t *TaskRequest) error {
    TODO()
    return errors.New("todo")
}
