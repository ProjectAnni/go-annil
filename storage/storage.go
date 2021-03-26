package storage

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type User struct {
	Username     string `json:"username"`
	RegisterDate int64  `json:"registerTime"`
	AllowShare   bool   `json:"allowShare"`
	IsAdmin      bool   `json:"isAdmin"`
}

type InviteCode struct {
	Code string `json:"code"`
	// -1 indicates infinity
	Limit string `json:"limit"`
}

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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS InviteCodes(\n    `Code` varchar(64) NOT NULL,\n    `Limit` int NOT NULL DEFAULT 0,\n    PRIMARY KEY(`Code`)\n)")
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

func ChangePassword(username, password string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE Users SET Password=? WHERE Username=?", hex.EncodeToString(h), username)
	return err
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

func ListUsers() []User {
	ret := make([]User, 0)
	rows, err := db.Query("SELECT Username,RegisterDate, AllowShare, IsAdmin FROM Users")
	if err != nil {
		return ret
	}
	for rows.Next() {
		var u User
		var t time.Time
		err = rows.Scan(&u.Username, &t, &u.AllowShare, &u.IsAdmin)
		if err != nil {
			return ret
		}
		u.RegisterDate = t.Unix()
		ret = append(ret, u)
	}
	return ret
}

func NewInviteCode(limit int) (string, error) {
	if limit == 0 {
		return "", fmt.Errorf("invalid limit")
	}
	code := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO InviteCodes(`Code`, `Limit`) VALUES (?,?)", code, limit)
	if err != nil {
		return "", err
	}
	return code, nil
}

func ListInviteCodes() []InviteCode {
	ret := make([]InviteCode, 0)
	rows, err := db.Query("SELECT Code, `Limit` FROM InviteCodes")
	if err != nil {
		return ret
	}
	for rows.Next() {
		var r InviteCode
		err = rows.Scan(&r.Code, &r.Limit)
		if err != nil {
			return ret
		}
		ret = append(ret, r)
	}
	return ret
}

func ShrinkInviteCode(code string) bool {
	ok := false
	err := db.QueryRow("SELECT EXISTS(SELECT * FROM InviteCodes WHERE Code=? AND (`Limit`>0 or `Limit`=-1))", code).Scan(&ok)
	if err != nil {
		return false
	}
	_, err = db.Exec("UPDATE InviteCodes SET `Limit`=`Limit`-1 WHERE Code=? AND `Limit`>0; DELETE FROM InviteCodes WHERE `Limit`=0", code)
	if err != nil {
		log.Printf("Failed to shrink invite code: %v", err)
	}
	return ok
}

func RevokeInviteCode(code string) error {
	_, err := db.Exec("DELETE FROM InviteCodes WHERE Code=?", code)
	return err
}
