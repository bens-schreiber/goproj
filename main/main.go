// Define all API endpoints and their algorithms
package main

import (
	"benschreiber.com/bres"
	"benschreiber.com/bsql"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"log"
)

func main() {

	log.SetPrefix("[main] ")
	log.SetFlags(log.Lmsgprefix)

	//Establish connection to local db using ./sqlConnector package
	bsql.Establishconnection()

	//Establish token pool
	bres.InitializeTokenMap()

	//Define API endpoints
	router := gin.Default()

	// Client endpoints
	client := "/api/client/"
	router.POST(client+"login", loginClient)
	router.POST(client+"register", registerClient)

	// Group endpoints
	group := "/api/group/"
	router.GET(group+":user", getGroup)
	router.POST(group+"create", postGroup)
	router.POST(group+"join", postGroupMember)
	router.POST(group+"coin", postCoin)

	//port 8080
	router.Run()
}

// METHOD: POST
// Generate API token in bres package
// Requires Username, Password headers
func loginClient(c *gin.Context) {

	// Validate headers exist
	// STATUS: 400 Bad Request on missing headers
	if !bres.ValidateHeaders(c, "Username", "Password") {
		return
	}

	//Grab headers
	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")

	// Validate userpass in allowed characters
	// STATUS: 400 bad request on illegal characters
	if !bres.ValidateUserPassRegex(c, user, pass) {
		return
	}

	// STATUS: 404 on nonexistant user
	if !bsql.UserExists(user) {
		c.AbortWithStatus(404)
		return
	}

	// Validate the credentials the user gave
	// STATUS: 401 Unauthorized on invalid credentials
	if !bsql.MatchUserPass(user, pass) {
		c.AbortWithStatus(401)
		return
	}

	// Create the token in memory, return in JSON
	// STATUS: 201 Created
	c.JSON(201, gin.H{"token": bres.AddClient(c.ClientIP(), user)})
}

// METHOD: POST
// Insert a new user into the database
// Requires Username, Password headers
func registerClient(c *gin.Context) {

	// Validate headers exist
	// STATUS: 400 Bad Request on missing headers
	if !bres.ValidateHeaders(c, "Username", "Password") {
		return
	}

	// Grab headers
	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")

	// Validate userpass in allowed characters
	// STATUS: 400 bad request on illegal characters
	if !bres.ValidateUserPassRegex(c, user, pass) {
		return
	}

	// Validate Username is unique
	// STATUS: 400 Bad Request on non unique user
	if bsql.UserExists(user) {
		log.Println("user already exists")
		c.AbortWithStatus(400)
		return
	}

	// Add user to db
	if _, err := bsql.InsertNewUser(user, pass); err != nil {
		log.Fatal(err)
	}

	// STATUS: 201 Created
	c.Status(201)
}

// METHOD: GET
// Return all Group fields and Group Members
// Requires Username, Token headers; user param
func getGroup(c *gin.Context) {

	// Validate userpass and Token fields exis
	// STATUS: 401 Unauthorized on invalid token
	// STATUS: 400 Bad Request on missing header; illegal chars
	// STATUS: 404 on non-existant user
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Grab user parameter
	user := c.Param("user")

	// Create return JSON
	// STATUS: 404 Not Found if user is not in a group
	group, ok := bsql.GetUserGroup(user)
	if !ok {
		c.AbortWithStatus(404)
		return
	}

	// STATUS: 200 OK
	c.JSON(200, group)
}

// METHOD: POST
// Insert a new group into the database
// Requires Username, Token headers
func postGroup(c *gin.Context) {

	// Validate userpass and Token fields exis
	// STATUS: 401 Unauthorized on invalid token
	// STATUS: 400 Bad Request on missing header; illegal chars
	// STATUS: 404 on non-existant user
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Grab user parameter
	user := c.GetHeader("Username")

	// Register new group
	if err := bsql.InsertNewGroup(user); err != nil {
		log.Fatal(err)
	}

	// STATUS: 200 OK
	c.Status(200)
}

// METHOD: POST
// Inserts user into a specified group
// Requires Username, Token, ID headers
func postGroupMember(c *gin.Context) {

	// Validate userpass and Token fields exis
	// STATUS: 401 Unauthorized on invalid token
	// STATUS: 400 Bad Request on missing header; illegal chars
	// STATUS: 404 on non-existant user
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Validate that ID is in the header
	// STATUS: 400 Bad Request on missing header
	if !bres.ValidateHeaders(c, "ID") {
		return
	}

	// Grab user and group id
	user := c.GetHeader("Username")
	id := c.GetHeader("ID")

	// STATUS: 404 Not Found on non-existant group
	if !bres.ValidateGroupExists(id) {
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

	// STATUS: 200, OK
	c.Status(200)

}

// METHOD: POST
// Updates group's coin
// Requires Username, Token, ID headers
func postCoin(c *gin.Context) {

	// Validate userpass and Token fields exis
	// STATUS: 401 Unauthorized on invalid token
	// STATUS: 400 Bad Request on missing header; illegal chars
	// STATUS: 404 on non-existant user
	if !bres.ValidateAuthentication(c) {
		return
	}

	// Validate that ID is in the header
	// STATUS: 400 Bad Request on missing header
	if !bres.ValidateHeaders(c, "ID") {
		return
	}

	// Grab user and group id
	user := c.GetHeader("Username")
	id := c.GetHeader("ID")

	// STATUS: 404 Not Found on non-existant group
	if !bres.ValidateGroupExists(id) {
		c.AbortWithStatus(404)
		return
	}

	// Validate the user is authorized to make a coin request
	// STATUS: 403 Forbidden on not high enough credentials
	if !bres.ValidateCoinRequest(c, user, id) {
		c.AbortWithStatus(403)
		return
	}

	// TODO: rewrite this753f04ed-d410-4b77-b091-8a0e547eefc1
	if err1, err2 := bsql.UpdateCoin(user, id); err1 != nil || err2 != nil {
		log.Fatal(err1)
		log.Fatal(err2)
	}

	// STATUS: 201 Created
	c.Status(201)

}
