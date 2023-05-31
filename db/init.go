package db

import (
	"net/http"
	"regexp"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var secretKey []byte
var database *gorm.DB = nil
var usernameSanitizer = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._]{1,30}[a-zA-Z0-9])?$`)
var passwordSanitizer = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*?]{8,128}$`)

const UsernameParameterString = "Username: 3-32 chars, alphanumerics & special chars: _."
const PasswordParameterString = "Password: 8-128 chars, alphanumerics & special chars: !@#$%^&*?"

type UpdateTaskRequest struct {
    NewTitle            string  `json:"new_title"`
    NewContents         string  `json:"new_contents"`
    ToggleCompleted     bool    `json:"toggle_completed"`
}

type CreateTaskRequest struct {
    Title       string  `json:"title" binding:"required"`
    Contents    string  `json:"contents"`
    Completed   bool    `json:"completed"`
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
    CreatedAt   time.Time   `gorm:"<-:create"`
    UserID      string      `gorm:"type:varchar(32);not null;<-:create"`
    Title       string      `gorm:"type:varchar(128);not null"`
    Completed   bool        `gorm:"type:bool;not null"`
    Contents    string      `gorm:"type:longtext;not null"`
}

type APIError struct {
    status      int
    message     string
}

func (e *APIError) Error() string {
    return e.message
}

func (e *APIError) HttpStatus() int {
    return e.status
}

func TODO() *APIError {
    return &APIError{ http.StatusNotImplemented, "Not implemented" }
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

func CreateUser(u *User) *APIError {
    if result := database.Create(u); result.Error != nil {
        return &APIError{ http.StatusFound, "Couldn't create user" }
    }
    return nil
}

func ReadUser(username string, body *User) *APIError {
    if result := database.Take(body, "username = ?", username); result.Error != nil {
        return &APIError{ http.StatusNotFound, "Couldn't find user" }
    }
    return nil
}

func UpdateUser(username string, new_hash []byte, new_salt []byte) *APIError {
    return TODO()
}

func DeleteUser(username string) *APIError {
    return TODO()
}

func defaultServerError() *APIError {
    return &APIError{ http.StatusInternalServerError, "Couldn't process request" }
}

func CreateTask(username string, t *CreateTaskRequest) *APIError {
    task := Task {
        UserID: username,
        Title: t.Title,
        Completed: t.Completed,
        Contents: t.Contents,
    }
    if result := database.Create(&task); result.Error != nil {
        return defaultServerError()
    }
    return nil
}

func ReadTask(username string, id int) (*TaskFullResponse, *APIError) {
    var task TaskFullResponse
    if result := database.Raw("SELECT id, title, contents, completed FROM tasks WHERE user_id = ? AND id = ? AND deleted_at IS NULL", username, id).Scan(&task); result.Error != nil {
        return nil, defaultServerError()
    } else if result.RowsAffected == 0 {
        return nil, &APIError{ http.StatusNotFound, "Couldn't find task" }
    }
    return &task, nil
}

func ReadAllTasks(username string) ([]TaskHalfResponse, *APIError) {
    tasks := []TaskHalfResponse{}
    if err := database.Raw("SELECT id, title, completed FROM tasks WHERE user_id = ? AND deleted_at IS NULL", username).Find(&tasks).Error; err != nil {
        return nil, defaultServerError()
    }
    return tasks, nil
}

func UpdateTask(username string, id int, t *UpdateTaskRequest) *APIError {
    var task Task
    if result := database.Raw("SELECT id, title, contents, completed FROM tasks WHERE user_id = ? AND id = ? AND deleted_at IS NULL", username, id).Scan(&task); result.Error != nil {
        return defaultServerError()
    } else if result.RowsAffected == 0 {
        // should we just create the task for them if it wasn't found?
        return &APIError{ http.StatusNotFound, "Couldn't find task" }
    }

    if t.NewTitle != "" {
        task.Title = t.NewTitle
    }

    if t.NewContents != "" {
        task.Contents = t.NewContents
    }

    if t.ToggleCompleted {
        task.Completed = !task.Completed
    }

    task.UpdatedAt = time.Now()

    if result := database.Save(&task); result.Error != nil {
        return defaultServerError()
    }

    return nil
}

func DeleteTask(username string, id int) *APIError {
    if results := database.Where("user_id = ? AND id = ?", username, id).Delete(&Task{}); results.Error != nil {
        return defaultServerError()
    } else if results.RowsAffected == 0 {
        return &APIError{ http.StatusNotFound, "Couldn't find task to delete" }
    }

    return nil
}
