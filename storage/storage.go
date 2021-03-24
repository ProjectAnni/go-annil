package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}
	// TODO
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Users (\n    `Username` varchar(64) NOT NULL,\n    `AllowShare` int NOT NULL DEFAULT 0,\n    `RegisterTime` datetime NOT NULL,\n    `Password` varchar(64) NOT NULL,\n    PRIMARY KEY (`Username`)\n)")
	if err != nil {
		return err
	}
	return nil
}
