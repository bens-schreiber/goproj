package bres

import (
	"benschreiber.com/bsql"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
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

func InitializeTokenMap() {
	configLogger()
	tokens = make(map[string]*Client)
	user_to_token = make(map[string]string)
}

func AddClient(ip string, username string) string {

	//UUID
	ret := uuid.New().String()

	// if the user already has a registerd token, delete it.
	if _, ok := user_to_token[username]; ok {
		delete(tokens, user_to_token[username])
		delete(user_to_token, username)
		log.Println("refreshing a users token")
	}

	//Add a new Client to the tokens map with a 6 hour from now expiration date
	tokens[ret] = &Client{
		IPAddress:  ip,
		Username:   username,
		Expiration: time.Now().Add(time.Hour * 6),
	}

	user_to_token[username] = ret

	log.Println("Tokens:", tokens)
	return ret

}

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

func ValidateAuthentication(c *gin.Context) bool {

	//Validate all headers are present in request
	if !ValidateHeaders(c, "Token", "Username") {
		return false
	}

	//Grab auth fields
	token := c.GetHeader("Token")
	username := c.GetHeader("Username")

	//Check if token exists and is valid
	if _, ok := tokens[token]; !ok {
		log.Println("invalid token")
		c.AbortWithStatus(401)
		return false
	}

	//Validate token fields
	client := tokens[token]
	if !client.Expiration.After(time.Now()) ||
		client.Username != username ||
		client.IPAddress != c.ClientIP() {
		log.Println("compromised token")
		//Remove the token because it is comprimised
		delete(tokens, token)
		c.AbortWithStatus(401)
		return false
	}
	return true
}

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

func configLogger() {
	log.SetPrefix("[bres] ")
	log.SetFlags(log.Lmsgprefix)
}
