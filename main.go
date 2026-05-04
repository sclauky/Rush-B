package main

import (

    "socialauth/backend/forum/auth/db"
    "socialauth/backend/forum/auth/handlers"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "log"
    "os"
    
)

func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading the .env file")
    }

    database := db.Open("./backend/database/data.db")
    defer database.Close()

    authHandler := &handlers.AuthHandler{DB: database}

    router := gin.New()
    router.Use(gin.Logger())

    router.POST("/register", func(c *gin.Context) {
        authHandler.Register(c.Writer, c.Request)
    })
    router.POST("/login", func(c *gin.Context) {
        authHandler.Login(c.Writer, c.Request)
    })

    router.GET("/api-1", func(c *gin.Context) {
        c.JSON(200, gin.H{"success": "Access granted for api-1"})
    })
    router.GET("/api-2", func(c *gin.Context) {
        c.JSON(200, gin.H{"success": "Access granted for api-2"})
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    router.Run(":" + port)
}