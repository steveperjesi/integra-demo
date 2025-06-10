package user

import "errors"

var (
	ErrMissingUserID    = errors.New("missing user_id")
	ErrMissingUserName  = errors.New("missing user_name")
	ErrMissingFirstName = errors.New("missing first_name")
	ErrMissingLastName  = errors.New("missing last_name")
	ErrMissingEmail     = errors.New("missing email")
	ErrUserNotFound     = errors.New("user not found")

	ErrUpdateUserMissingValues = errors.New("no values to update")
	ErrUpdateUserNoRows        = errors.New("no rows updated")

	ErrUserExists = errors.New("user_name already exists")
)
