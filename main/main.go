package main

import (
//	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	//"encoding/json"
	"benschreiber.com/bsql"
	"regexp"
	"log"
	"database/sql"
)

func main() {

	//Establish connection to local db using ./sqlConnector package
	bsql.Establishconnection()
	
	//Define API endpoints
	router := gin.Default()
	router.GET("/api/group/:user", getGroup)	


	//port 8080
	router.Run()
}

//TODO: Setup token system, HTTPS
func getGroup(c *gin.Context) {

	//if !validateAuthentication(c) { return }

	//Grab user parameter
	user := c.Param("user")

	//Handle a bad username that contains illegal characters
	regex, _ := regexp.Compile("[^A-Za-z0-9]+")
	if regex.MatchString(user) {
		c.JSON(400, gin.H{"":""})
		fmt.Println("bad req")
		return
	}

	//Create a Group SQL table struct to return on ACCEPTED
	var query bsql.Group
	err := bsql.QueryDB(user,
	"select * from _group where _group.id=(select group_id from group_member where username=?)",
	&query.ID, &query.Token, &query.Creator)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"":""})
			return
		} else { log.Fatal(err) }
	}

	//Accept request, return the query result.
	c.JSON(200, query)
	fmt.Println(query)	

}
