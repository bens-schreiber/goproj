package main

import (
	"benschreiber.com/bres"
	"benschreiber.com/bsql"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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
	router.POST("/api/group/join", postGroupMember)
	router.POST("/api/group/coin", postCoin)

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

	// Verify the user exists
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// Register new group
	if err := bsql.InsertNewGroup(user); err != nil {
		log.Fatal(err)
	}

	c.Status(200)

}

// POST
func postGroupMember(c *gin.Context) {
	// Aborts on invalid auth or headers (token and username are in all requests)
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Validate that ID is in the header
	if !bres.ValidateHeaders(c, "ID") {
		return
	}

	// Grab user and group id
	user := c.GetHeader("Username")
	id := c.GetHeader("ID")

	// Verify the user exists
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// insert group member to group. if err, see if duplicate entry (group mems + id must be unique)
	if err := bsql.InsertGroupMember(user, id); err != nil {
		if _, ok := err.(*mysql.MySQLError); !ok {
			log.Fatal(err)
		}
		if err.(*mysql.MySQLError).Number == 1062 {
			log.Println("User already in group they tried to join")
			c.AbortWithStatus(400)
			return
		}
	}

}

// POST
func postCoin(c *gin.Context) {
	// Aborts on invalid auth or headers (token and username are in all requests)
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Validate that ID is in the header
	if !bres.ValidateHeaders(c, "ID") {
		return
	}

	// Grab user and group id
	user := c.GetHeader("Username")
	id := c.GetHeader("ID")

	// Verify the user exists
	if !bsql.ValidateUserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// Return a Forbidden status code if someone who isnt the coin holder makes request
	if !bres.ValidateCoinRequest(c, user, id) {
		c.AbortWithStatus(403)
		return
	}

	if err1, err2 := bsql.UpdateCoin(user, id); err1 != nil || err2 != nil {
		log.Fatal(err1)
		log.Fatal(err2)
	}

	c.Status(201)

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
