package bsql

import (
	"database/sql"
	//"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

//SQL table structs
type Group struct {
	ID      string `json:"id"`
	Token   int    `json:"token"`
	Creator string `json:"creator"`
	TokenHolder string `json:"token_holder"`
}

type GroupMember struct {
	GroupID  string `json:"group_id"`
	Username string `json:"username"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//TODO: multiple params???!?!?!?!?
func QueryDB(param string, query string, args... interface{}) (error) {

	//Use a PreparedStatement to run query
	stmt, err := db.Prepare(query)
	if err != nil {log.Fatal(err)}

	return  stmt.QueryRow(param).Scan(args...)
}

func Establishconnection() {

	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "root",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "puapp",
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	
	fmt.Println("Connected!")
}

func main() {}
