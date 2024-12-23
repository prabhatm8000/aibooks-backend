package auth

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"example/aibooks-backend/errorHandling"
	"example/aibooks-backend/models/otps"
	"example/aibooks-backend/models/users"
	"example/aibooks-backend/utils"
)

func CreateAccount(c *gin.Context) {
	var data struct {
		FirstName       string `json:"first_name" binding:"required"`
		LastName        string `json:"last_name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Otp             string `json:"otp" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,min=8,eqfield=Password"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println(err)
		c.IndentedJSON(400, gin.H{"message": "Invalid request"})
		return
	}

	if !otps.CompareOtp(data.Email, data.Otp) {
		c.IndentedJSON(400, gin.H{"message": "Invalid OTP"})
		return
	}

	bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Failed to create account"})
		return
	}

	user := users.Users{
		FirstName:     data.FirstName,
		LastName:      data.LastName,
		Email:         data.Email,
		Password:      string(bcryptPassword),
		EmailVerified: true,
	}
	userId, err := users.AddUser(user)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Failed to create account"})
		return
	}

	createJWTTokenCookie(c, userId.Hex())
	c.IndentedJSON(201, gin.H{"message": "Account created successfully"})
}

func Login(c *gin.Context) {
	var data struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	existingUser, err := users.GetUserByEmail(data.Email)
	if err != nil {
		c.IndentedJSON(404, gin.H{"message": "User not found."})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(data.Password))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Invalid credentials."})
		return
	}

	createJWTTokenCookie(c, existingUser.Id.Hex())
	c.IndentedJSON(200, gin.H{
		"message": "Success",
	})
}

func createJWTTokenCookie(c *gin.Context, userId string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Failed to generate token"})
		return
	}

	secure := os.Getenv("ENV") == "PROD"

	c.Header("Set-Cookie", fmt.Sprintf("auth-token=%s; SameSite=None; Secure=%v; Path=/", tokenString, secure))
}

func Logout(c *gin.Context) {
	c.Header("Set-Cookie", fmt.Sprintf("auth-token=%s; SameSite=None; Secure=%v; Path=/", "", false))
	c.IndentedJSON(200, gin.H{"message": "Successfully logged out"})
}

func GetUserDetails(c *gin.Context) {
	userId := c.GetString("user_id")

	user, err := users.GetUserById(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
	}

	c.IndentedJSON(200, user)
}

func SendOtp(c *gin.Context) {
	var data struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	otp, err := otps.GenerateAndSaveOtpFor(data.Email)
	if err == errorHandling.ErrTooSoon {
		c.IndentedJSON(400, gin.H{"message": "Too soon, please try again later."})
		return
	} else if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong generating OTP."})
		return
	}

	err = utils.SendOtpEmail(data.Email, otp)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong sending OTP."})
		return
	}

	c.IndentedJSON(200, gin.H{"message": "OTP sent successfully"})
}
