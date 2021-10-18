// Rest Endpoint Security helper function package
// Protects public API by validating requests
package bres

import (
	"benschreiber.com/bsql"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"regexp"
	"time"
)

type Client struct {
	IPAddress  string
	Username   string
	Expiration time.Time
}

func (c Client) String() string {
	return fmt.Sprintf("[%s, %s]", c.IPAddress, c.Username)

}

// Maps in memory that retain active tokens
var tokens map[string]*Client
var user_to_token map[string]string

// Initialize maps in memory
func InitializeTokenMap() {
	configLogger()
	tokens = make(map[string]*Client)
	user_to_token = make(map[string]string)
}

// Insert a new token into memory
func AddClient(ip string, username string) string {

	// UUID
	ret := uuid.New().String()

	// if the user already has a registerd token, delete it.
	if _, ok := user_to_token[username]; ok {
		delete(tokens, user_to_token[username])
		delete(user_to_token, username)
		log.Println("refreshing a users token")
	}

	// Add a new Client to the tokens map with a 6 hour from now expiration date
	tokens[ret] = &Client{
		IPAddress:  ip,
		Username:   username,
		Expiration: time.Now().Add(time.Hour * 6),
	}

	user_to_token[username] = ret

	log.Println("Tokens:", tokens)
	return ret

}

// Checks context for specified headers
// STATUS: 400 Bad Request on missing header
func ValidateHeaders(c *gin.Context, args ...string) bool {
	for _, v := range args {
		if c.GetHeader(v) == "" {
			log.Println("invalid or missing headers")
			c.AbortWithStatus(400)
			return false
		}
	}
	return true
}

// General authentication validation
// Validate API Tokens
func ValidateAuthentication(c *gin.Context) bool {

	// Validate all headers are present in request
	if !ValidateHeaders(c, "Token", "Username") {
		return false
	}

	// Grab auth fields
	token := c.GetHeader("Token")
	username := c.GetHeader("Username")

	// Validate user is in allowed characters
	// STATUS: 400 Bad Request on illegal characters
	if !ValidateUserPassRegex(c, username, "") {
		return false
	}

	// Verify the user exists
	// STATUS: 404 Not Found on non-existant user
	if !bsql.UserExists(username) {
		c.AbortWithStatus(404)
		return false
	}

	// Check if api token exists
	// STATUS: 401 Unauthorized on invalid token
	if _, ok := tokens[token]; !ok {
		log.Println("invalid token")
		c.AbortWithStatus(401)
		return false
	}

	// Validate token field
	// STATUS: 401 Unauthorized on invalid token
	client := tokens[token]
	if !client.Expiration.After(time.Now()) ||
		client.Username != username ||
		client.IPAddress != c.ClientIP() {

		log.Println("compromised, expired or invalid")
		// Remove invalidated Token
		delete(tokens, token)
		c.AbortWithStatus(401)
		return false

	}
	return true
}

func ValidateGroupExists(id string) bool {
	return bsql.GroupExists(id)
}


// Check if user is capable of making a coin request
func ValidateCoinRequest(c *gin.Context, user string, id string) bool {
	err := bsql.SelectCoinHolder(user, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false

		}
		log.Fatal(err)
	}

	return true

}

func ValidateUserPassRegex(c *gin.Context, username string, password string) bool {

	// Handle a bad username that contains illegal characters
	if regex, _ := regexp.Compile("[^A-Za-z0-9]+"); regex.MatchString(username) {
		log.Println("username does not follow guidelines")
		c.AbortWithStatus(400)
		return false
	}

	//See if password contains any whitespaces
	if password != "" {
		if regex, _ := regexp.Compile("\\s+"); regex.MatchString(password) {
			log.Println("password does not follow guidelines")
			c.AbortWithStatus(400)
			return false
		}
	}

	return true
}

func configLogger() {
	log.SetPrefix("[bres] ")
	log.SetFlags(log.Lmsgprefix)
}
