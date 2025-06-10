package db

import "database/sql"

type UserDB struct {
	UserID     int64
	UserName   string
	FirstName  string
	LastName   string
	Email      string
	UserStatus string
	Department sql.NullString
}
