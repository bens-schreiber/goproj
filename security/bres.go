package bres

import (
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

var tokens map[string]*Client

func InitializeTokenMap() {
	configLogger()
	tokens = make(map[string]*Client)
}

func AddClient(ip string, username string) string {

	//UUID
	ret := uuid.New().String()

	//Add a new Client to the tokens map with a 6 hour from now expiration date
	tokens[ret] = &Client{
		IPAddress:  ip,
		Username:   username,
		Expiration: time.Now().Add(time.Hour * 6),
	}
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

func configLogger() {
	log.SetPrefix("[bres] ")
	log.SetFlags(log.Lmsgprefix)
}
