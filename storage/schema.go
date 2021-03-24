package storage

import "time"

type User struct {
	Username     string `json:"username"`
	AllowShare   bool   `json:"allowShare"`
	RegisterDate time.Time
}
