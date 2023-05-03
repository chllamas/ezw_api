package main 

import ( 
    "os"
    "log"
    "strings"
    "net/http"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

func handle_pong(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "pong",
    })
}

func AuthMiddleware(clerkClient *clerk.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        accTok := c.Request.Header.Get("Authorization")
        tokArr := strings.Split(accTok, " ")
        accessToken := tokArr[1]
        log.Printf("Recieved token: %s\n", accessToken)

        if accessToken == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        if _, err := (*clerkClient).VerifyToken(accessToken); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func main() {
    clerkClient, err := clerk.NewClient(os.Getenv("CLERK_SECRET_KEY"))

    if err != nil {
        log.Fatalf("Clerk Client error: %v", err)
    }

    router := gin.Default()
    router.Use(AuthMiddleware(&clerkClient))
    router.GET("/ping", handle_pong)
    router.Run("127.0.0.1:8000")
}

