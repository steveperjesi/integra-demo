package user

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/steveperjesi/integra-demo/internal/db"
)

const (
	DbName = "users"
)

func (u *User) SetUserStatus(status string) {
	switch strings.ToLower(status) {
	case "a":
		// Active
		u.setUserStatusActive()
	case "i":
		// Inactive
		u.setUserStatusInactive()
	case "t":
		// Terminated
		u.setUserStatusTerminated()
	default:
		// Default to inactive
		u.setUserStatusInactive()
	}
}

func (u *User) setUserStatusActive() {
	u.UserStatus = "A"
}

func (u *User) setUserStatusInactive() {
	u.UserStatus = "I"
}

func (u *User) setUserStatusTerminated() {
	u.UserStatus = "T"
}

func (req *User) ValidateNewUserRequest() error {
	if req.UserName == "" {
		return ErrMissingUserName
	}

	if req.FirstName == "" {
		return ErrMissingFirstName
	}

	if req.LastName == "" {
		return ErrMissingLastName
	}

	if req.Email == "" {
		return ErrMissingEmail
	}

	// Department is optional and allowed to be empty

	// Verify the status, defaulting to `inactive`
	req.SetUserStatus(req.UserStatus)

	return nil
}

func ValidateUserID(input string) (int64, error) {
	id, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user_id: must be an integer")
	}
	return id, nil
}

func (u *User) ConvertToUserDB() db.UserDB {
	userDB := db.UserDB{
		UserID:     u.ID,
		UserName:   u.UserName,
		UserStatus: u.UserStatus,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Email:      u.Email,
		Department: sql.NullString{},
	}

	if u.Department != nil {
		userDB.Department = sql.NullString{
			Valid:  true,
			String: *u.Department,
		}
	}

	return userDB
}

func ConvertToUser(udb *db.UserDB) User {
	user := User{
		ID:         udb.UserID,
		UserName:   udb.UserName,
		UserStatus: udb.UserStatus,
		FirstName:  udb.FirstName,
		LastName:   udb.LastName,
		Email:      udb.Email,
	}

	if udb.Department.Valid {
		user.Department = &udb.Department.String
	}

	return user
}
