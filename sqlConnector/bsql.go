package bsql

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"log"
)

var db *sql.DB

//SQL table structs
type Group struct {
	ID          string   `json:"id"`
	Token       int      `json:"token"`
	Creator     string   `json:"creator"`
	TokenHolder string   `json:"token_holder"`
	Members     []string `json:"members"`
}

type GroupMember struct {
	GroupID  string `json:"group_id"`
	Username string `json:"username"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var insertUserQuery *sql.Stmt
func InsertNewUser(user string, pass string) (sql.Result, error) {
	return insertUserQuery.Exec(user, pass)
}

var userExistsQuery *sql.Stmt
func ValidateUserExists(user string) bool {
	//Temp fix, .Err() does not seem to return ErrNoRows properly
	var username string
	err := userExistsQuery.QueryRow(user).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("user does not exist")
			return false
		}
		log.Fatal(err)
	}

	return true
}

var matchUserPassQuery *sql.Stmt
func ValidateCredentials(user string, pass string) bool {
	var username string
	err := matchUserPassQuery.QueryRow(user, pass).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Credentials invalid")
			return false
		}
		log.Fatal(err)
	}

	return true
}

var userGroupQuery *sql.Stmt
var groupUsersQuery *sql.Stmt
func GetUserGroup(user string) (*Group, bool) {
	var group Group

	err := userGroupQuery.QueryRow(user).Scan(
		&group.ID,
		&group.Token,
		&group.Creator,
		&group.TokenHolder)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("group not found")
			return nil, false
		}
		log.Fatal(err)
	}

	rows, err := groupUsersQuery.Query(group.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		rows.Scan(&username)
		group.Members = append(group.Members, username)
	}

	return &group, true
}

var insertGroupQuery *sql.Stmt
var insertGroupMemberQuery *sql.Stmt
func InsertNewGroup(user string) (error) {

	// group id
	id := uuid.New().String()
	tokenDefaultValue := 1

	{_, err := insertGroupQuery.Exec(id, tokenDefaultValue, user, user)
	if err != nil { return err }}
	
	_, err := insertGroupMemberQuery.Exec(id, user)
	if err != nil { return err }

	return nil
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

	setupPrepStates()
	configLogger()
	log.Println("Connected to Database!")

}

//Setup all prepared statements
func setupPrepStates() {
	var err error

	insertUserQuery, err = db.Prepare("insert into user(username, password) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	userExistsQuery, err = db.Prepare("select username from user where username=?")
	if err != nil {
		log.Fatal(err)
	}

	matchUserPassQuery, err = db.Prepare("select username from user where username=? and password=?")
	if err != nil {
		log.Fatal(err)
	}

	userGroupQuery, err = db.Prepare("select * from _group where _group.id=(select group_id from group_member where username=?)")
	if err != nil {
		log.Fatal(err)
	}

	groupUsersQuery, err = db.Prepare("select username from group_member where group_id=?")
	if err != nil {
		log.Fatal(err)
	}

	insertGroupQuery, err = db.Prepare("insert into _group(id, token, creator, token_holder) values (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	insertGroupMemberQuery, err = db.Prepare("insert into group_member(group_id, username) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
}

func configLogger() {
	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix("[bsql] ")
}
