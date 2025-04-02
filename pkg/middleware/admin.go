package middleware

import (
	"fmt"
	"log"
	"net/http"
	"postui_api/pkg/auth"
	"postui_api/pkg/models"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		userStr, ok := username.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username format"})
			c.Abort()
			return
		}
		if exists {
			fmt.Println(userStr) // Type assert to string
			if userStr == "admin" {
				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized, this function is only available for admin user.",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized, username don't exists",
			})
			c.Abort()
			return
		}
	}
}

func CreateAdmin(db *gorm.DB) {
	var user models.LoginUser

	user.Username = "admin"
	user.Password = "P@$$w0rd"

	// Hash the password
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Fatal("Cannot hash admin password")
	}

	// Create new user
	newUser := models.User{Username: user.Username, Password: hashedPassword}

	// Save the user to the database
	db.Create(&newUser)
}
