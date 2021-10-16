package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"benschreiber.com/bsql"
	"benschreiber.com/bres"
	"regexp"
	"log"
	"database/sql"
)

//TODO: setup ratelimiting
func main() {

	//Establish connection to local db using ./sqlConnector package
	bsql.Establishconnection()
	
	//Establish token pool
	bres.InitializeTokenMap()

	//Define API endpoints
	router := gin.Default()
	router.GET("/api/group/:user", getGroup)	
	router.POST("/api/client/login", loginClient)
	router.POST("/api/client/register", registerClient)

	//port 8080
	router.Run()
}

//Setup the insert query
func registerClient(c *gin.Context) {
	if !bres.ValidateHeaders(c, "Username", "Password") { return }

	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")

	//Validate that the password and username are within the allowed characters
	if !validateUserPassRegex(c, user, pass) { return }
	
	//var Query bsql.User
	//err :- bsql.QueryDB(

}


func loginClient(c *gin.Context) {

	//Aborts on invalid headers
	if !bres.ValidateHeaders(c, "Username", "Password") { return }
 
	user := c.GetHeader("Username")
	pass := c.GetHeader("Password")
	
	//Validate that the password and username are within the allowed characters
	if !validateUserPassRegex(c, user, pass) { return }


	{
		var query bsql.User
		err := bsql.QueryDB(user, 
		"select * from user where username=?",
		&query.Username, &query.Password)

		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatus(404)
				return
			}
		}
	}

	var query bsql.User
	err := bsql.QueryDB(pass, 
	fmt.Sprintf("select * from user where username='%s' and password=?", user),
	&query.Username, &query.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(401)
			return
		}
	}

	c.JSON(201, gin.H{"token":bres.AddToken(c.ClientIP(), user )})
}


func getGroup(c *gin.Context) {

	//Aborts on invalid auth or headers (token and username are in all requests)
	if !bres.ValidateAuthentication(c) { return }

	//Grab user parameter
	user := c.Param("user")

	//Handle a bad username that contains illegal characters
	regex, _ := regexp.Compile("[^A-Za-z0-9]+")
	if regex.MatchString(user) {
		c.AbortWithStatus(400)
		return
	}

	//Create a Group SQL table struct to return on ACCEPTED
	var query bsql.Group
	err := bsql.QueryDB(user,
	"select * from _group where _group.id=(select group_id from group_member where username=?)",
	&query.ID, &query.Token, &query.Creator, &query.TokenHolder)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		} else { log.Fatal(err) }
	}

	//Accept request, return the query result.
	c.JSON(200, query)
}

func validateUserPassRegex(c *gin.Context, username string, password string) bool {
	//Handle a bad username that contains illegal characters
	if regex, _ := regexp.Compile("[^A-Za-z0-9]+"); regex.MatchString(username) {
		c.AbortWithStatus(400)
		return false
	}

	//See if password contains any whitespaces
	if regex, _ := regexp.Compile("\\s+"); regex.MatchString(password) {
		c.AbortWithStatus(400)
		return false
	}

	return true
}
