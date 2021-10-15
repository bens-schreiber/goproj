package bres

import (
	"github.com/gin-gonic/gin"
	"time"
	"github.com/google/uuid"
)

type Client struct {
	IPAddress string
	Username string
	Expiration time.Time
}

var tokens map[string]Client

func InitializeTokenMap() {
	tokens = make(map[string]Client)
}

func AddToken(ip string, username string) string {

	//UUID
	ret := uuid.New().String()

	//Add a new Client to the tokens map with a 6 hour from now expiration date
	tokens[ret] = Client{
	IPAddress: ip,
	Username: username,
	Expiration: time.Now().Add(time.Hour * 6),
	}
	return ret

}

func validateHeaders(c *gin.Context, args... string) bool {
	for _, v := range args {
		if c.GetHeader(v) == "" { 
			c.AbortWithStatus(400)
			return false
		}
	}
	return true
}


//TODO: Get IP Addresses working
func ValidateAuthentication(c *gin.Context) bool {
	
	//Validate all headers are present in request
	if !validateHeaders(c, "Token", "Username", "X-FORWARDED-FOR") { return false }

	//Grab auth fields
	token := c.GetHeader("Token")
	username := c.GetHeader("Username")
	
	//Check if token exists and is valid
	if _, ok := tokens[token]; !ok {
		c.AbortWithStatus(401)
		return false
	} 
	
	//Validate token fields
	client := tokens[token]
	if client.Expiration.After(time.Now()) ||
	client.Username != username ||
	client.IPAddress != c.GetHeader("X-FORWARDED-FOR") {
		//Remove the token because it is comprimised
		delete(tokens, token)
		c.AbortWithStatus(401)
		return false
	}
	return true
}
