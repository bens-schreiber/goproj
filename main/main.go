package main

import (
	"benschreiber.com/bres"
	"benschreiber.com/bsql"
	"github.com/gin-gonic/gin"
	"log"
	"regexp"
)

//TODO: setup ratelimiting
func main() {

	log.SetPrefix("[main] ")
	log.SetFlags(log.Lmsgprefix)

	//Establish connection to local db using ./sqlConnector package
	bsql.Establishconnection()

	//Establish token pool
	bres.InitializeTokenMap()

	//Define API endpoints
	router := gin.Default()
	router.GET("/api/group/:user", getGroup)
	router.POST("/api/client/login", loginClient)
	router.POST("/api/client/register", registerClient)
	router.POST("/api/group/create", postGroup)

	//port 8080
	router.Run()
}


// POST
func registerClient(c *gin.Context) {

	//Validate headers exist
	if !bres.ValidateHeaders(c, "Username", "Password") {
		return
	}

	//Grab headers
	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")

	//Validate that the password and username are within the allowed characters
	if !validateUserPassRegex(c, user, pass) {
		return
	}

	//If a user exists with that username, return with Bad Request
	if bsql.ValidateUserExists(user) {
		log.Println("user already exists")
		c.AbortWithStatus(400)
		return
	}

	//Add user to db
	if _, err := bsql.InsertNewUser(user, pass); err != nil {
		log.Fatal(err)
	}

	// Return 201 created code
	c.JSON(201, "")
}

// POST
func loginClient(c *gin.Context) {

	// Validate headers exist
	if !bres.ValidateHeaders(c, "Username", "Password") {
		return
	}

	//Grab headers
	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")

	// Validate that the password and username are within the allowed characters
	if !validateUserPassRegex(c, user, pass) {
		return
	}

	// If a user with that username isnt found, return 404
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// Validate the credentials the user gave
	if !bsql.ValidateCredentials(user, pass) {
		c.AbortWithStatus(401)
		return
	}

	c.JSON(201, gin.H{"token": bres.AddClient(c.ClientIP(), user)})
}

// POST
func postGroup(c *gin.Context) {

	// Aborts on invalid auth or headers (token and username are in all requests)
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Grab user parameter
	user := c.GetHeader("Username")

	log.Println(user)

	// Verify the user exists
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}
	
	// Register new group
	if err := bsql.InsertNewGroup(user); err != nil {
		log.Fatal(err)
	}

	c.JSON(200, "")

}



// GET
func getGroup(c *gin.Context) {

	// Aborts on invalid auth or headers (token and username are in all requests)
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Grab user parameter
	user := c.Param("user")

	// Handle a bad username that contains illegal characters
	if !validateUserPassRegex(c, user, "") {
		return
	}

	// Verify the user exists
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// Create group struct
	group, ok := bsql.GetUserGroup(user)
	if !ok {
		c.AbortWithStatus(404)
		return
	}

	c.JSON(200, group)
}

func validateUserPassRegex(c *gin.Context, username string, password string) bool {
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
