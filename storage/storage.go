package storage

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var db *sql.DB

func Init() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Users (\n    `Username` varchar(64) NOT NULL,\n    `AllowShare` int NOT NULL DEFAULT 0,\n    `RegisterTime` datetime NOT NULL DEFAULT (DATETIME('now')),\n    `Password` varchar(64) NOT NULL,\n    `Admin` int NOT NULL DEFAULT 0,\n    PRIMARY KEY (`Username`)\n)")
	if err != nil {
		return err
	}
	// If there are no users, add a default admin account
	// Default password is "12345"
	_, err = db.Exec("INSERT INTO Users(Username, Password, `Admin`) SELECT \"Admin\", \"24326124313024645a41414b316b7045334557356c56587a7549537165754a773271676f555063794e754a49396e6c62566a35385a5142526b654f61\", 1 WHERE NOT EXISTS(SELECT * FROM Users)")
	if err != nil {
		return err
	}
	return nil
}

func Register(username, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO Users(Username,Password) values (?,?)", username, hex.EncodeToString(hashed))
	if err != nil {
		log.Printf("Failed to register user %s: %v\n", username, err)
	}
	return err
}

func CheckPassword(username, password string) bool {
	h, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Println(hex.EncodeToString(h))
	var hashed string
	err := db.QueryRow("SELECT `Password` FROM Users WHERE Username=?", username).Scan(&hashed)
	if err != nil {
		return false
	}
	pwd, err := hex.DecodeString(hashed)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(pwd, []byte(password))
	return err == nil
}

func RegisterDate(username string) (time.Time, error) {
	var t time.Time
	err := db.QueryRow("SELECT `RegisterTime` FROM Users WHERE Username=?", username).Scan(&t)
	return t, err
}

func RevokeUser(username string) (err error) {
	_, err = db.Exec("DELETE FROM Users WHERE Username=?", username)
	return
}

func UserExists(username string) bool {
	ret := false
	_ = db.QueryRow("SELECT EXISTS(SELECT * FROM Users WHERE Username=?)", username).Scan(&ret)
	return ret
}

func SetAllowShare(username string, allow bool) error {
	_, err := db.Exec("UPDATE Users SET AllowShare=? WHERE Username=?", allow, username)
	if err != nil {
		log.Printf("Failed to set allow share for %s: %v\n", username, err)
	}
	return err
}

func AllowShare(username string) bool {
	ret := false
	_ = db.QueryRow("SELECT AllowShare FROM Users WHERE Username=?", username).Scan(&ret)
	return ret
}

func IsAdmin(username string) bool {
	ret := false
	_ = db.QueryRow("SELECT `Admin` FROM Users WHERE Username=?", username).Scan(&ret)
	return ret
}

func SetAdmin(username string, admin bool) error {
	_, err := db.Exec("UPDATE Users SET `Admin`=? WHERE Username=?", admin, username)
	if err != nil {
		log.Printf("Failed to set allow share for %s: %v\n", username, err)
	}
	return err
}
