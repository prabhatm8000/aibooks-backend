package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"example/aibooks-backend/authenticator"
	"example/aibooks-backend/models/users"
)

func LoginHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := generateRandomState()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// getting currentUrl from client
		var currentUrl struct {
			CurrentUrl string `json:"currentUrl"`
		}
		if err := c.ShouldBindJSON(&currentUrl); err != nil {
			c.IndentedJSON(400, gin.H{"message": "Something went wrong while getting currentUrl"})
			return
		}

		session := sessions.Default(c)
		session.Set("state", state)
		session.Set("currentUrl", currentUrl.CurrentUrl)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		// c.Redirect(http.StatusTemporaryRedirect, auth.AuthCodeURL(state))
		c.IndentedJSON(200, gin.H{"authUrl": auth.AuthCodeURL(state)})
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)
	return state, nil
}

func CallbackHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if c.Query("state") != session.Get("state") {
			c.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(c.Request.Context(), c.Query("code"))
		if err != nil {
			c.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profileClaims map[string]interface{}
		if err := idToken.Claims(&profileClaims); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		profileJSON, err := json.Marshal(profileClaims)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		updatedAtStr := profileClaims["updated_at"].(string)
		updatedAtTime, err := time.Parse(time.RFC3339, updatedAtStr) // Assuming `updated_at` is in RFC3339 format
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}

		updatedAtPrimitive := primitive.NewDateTimeFromTime(updatedAtTime)

		userId, err := users.AddUser(users.Users{
			Sub:           profileClaims["sub"].(string),
			Name:          profileClaims["name"].(string),
			GivenName:     profileClaims["given_name"].(string),
			FamilyName:    profileClaims["family_name"].(string),
			Nickname:      profileClaims["nickname"].(string),
			Email:         profileClaims["email"].(string),
			EmailVerified: profileClaims["email_verified"].(bool),
			Picture:       profileClaims["picture"].(string),
			Aud:           profileClaims["aud"].(string),
			Iss:           profileClaims["iss"].(string),
			Exp:           int64(profileClaims["exp"].(float64)),
			Iat:           int64(profileClaims["iat"].(float64)),
			Sid:           profileClaims["sid"].(string),
			UpdatedAt:     updatedAtPrimitive,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profileJSON)
		session.Set("user_id", userId.Hex())
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		var returnTo string = session.Get("currentUrl").(string)

		if returnTo == "" {
			ginMode := os.Getenv("GIN_MODE")
			switch ginMode {
			case "release":
				returnTo = os.Getenv("FRONTEND_PROD_URL")
			default:
				returnTo = os.Getenv("FRONTEND_DEV_URL")
			}
		}
		fmt.Println(returnTo)
		c.Redirect(http.StatusTemporaryRedirect, returnTo)
	}
}

func UserProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	profile := session.Get("profile")

	if profile == nil {
		c.IndentedJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	var profileJson map[string]interface{}
	if err := json.Unmarshal(profile.([]byte), &profileJson); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(200, profileJson)
}

func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	ginMode := os.Getenv("GIN_MODE")
	var returnTo string
	switch ginMode {
	case "release":
		returnTo = os.Getenv("FRONTEND_PROD_URL")
	default:
		returnTo = os.Getenv("FRONTEND_DEV_URL")
	}

	session.Clear()
	if err := session.Save(); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo)
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	c.IndentedJSON(200, gin.H{"authUrl": logoutUrl.String()})
	// c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}
