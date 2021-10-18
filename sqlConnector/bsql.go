// Helper function package for executing mysql queries
package bsql

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"log"
)

// SQL Database pointer
var db *sql.DB

// SQL: table _group
type Group struct {
	ID          string   `json:"id"`
	Token       int      `json:"coin"`
	Creator     string   `json:"creator"`
	TokenHolder string   `json:"coin_holder"`
	Members     []string `json:"members"`
}

// SQL: table group_member
type GroupMember struct {
	GroupID  string `json:"group_id"`
	Username string `json:"username"`
}

// SQL: table user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var insertUserQuery *sql.Stmt

func InsertNewUser(user string, pass string) (sql.Result, error) {
	return insertUserQuery.Exec(user, pass)
}


var groupExistsQuery *sql.Stmt

func GroupExists(id string) bool {

	// Return a value into group if group exists
	var group string
	err := groupExistsQuery.QueryRow(group).Scan(&group)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("group does not exist")
			return false
		}
		log.Fatal(err)
	}
	
	return true
}


var userExistsQuery *sql.Stmt

func UserExists(user string) bool {

	// Return a value into username if the user exists
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

func MatchUserPass(user string, pass string) bool {
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


var insertGroupMemberQuery *sql.Stmt

func InsertGroupMember(user string, id string) error {

	_, err := insertGroupMemberQuery.Exec(id, user)
	if err != nil {
		return err
	}

	return nil
}


var insertGroupQuery *sql.Stmt

func InsertNewGroup(user string) error {

	// group id
	id := uuid.New().String()
	tokenDefaultValue := 1

	_, err := insertGroupQuery.Exec(id, tokenDefaultValue, user, user)
	if err != nil {
		return err
	}

	if err2 := InsertGroupMember(user, id); err2 != nil {
		return err2
	}

	return nil
}


var selectCoinHolderQuery *sql.Stmt

func SelectCoinHolder(user string, id string) error {
	var username string
	return selectCoinHolderQuery.QueryRow(user, id).Scan(&username)
}


var updateCoinQuery *sql.Stmt
var updateCoinHolderQuery *sql.Stmt

func UpdateCoin(user string, id string) (error, error) {
	_, err1 := updateCoinQuery.Exec(id, user)
	_, err2 := updateCoinHolderQuery.Exec(id, id)
	return err1, err2

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

	insertGroupQuery, err = db.Prepare("insert into _group(id, coin, creator, coin_holder) values (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	insertGroupMemberQuery, err = db.Prepare("insert into group_member(group_id, username) values (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	selectCoinHolderQuery, err = db.Prepare("select coin_holder from _group where coin_holder=? and id=?")
	if err != nil {
		log.Fatal(err)
	}

	updateCoinQuery, err = db.Prepare("update _group set _group.coin = (_group.coin + 1) where id=? and coin_holder=?")
	if err != nil {
		log.Fatal(err)
	}

	updateCoinHolderQuery, err = db.Prepare("update _group set coin_holder=(select username from group_member where group_id=? order by rand() limit 1) where id=?")
	if err != nil {
		log.Fatal(err)
	}

	groupExistsQuery, err = db.Prepare("select id from _group where _group.id=?")
	if err != nil {
		log.Fatal(err)
	}

}

func configLogger() {
	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix("[bsql] ")
}
