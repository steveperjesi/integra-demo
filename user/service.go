package user

import (
	"database/sql"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID         int64   `json:"user_id"`
	UserName   string  `json:"user_name"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email"`
	UserStatus string  `json:"user_status"`
	Department *string `json:"department,omitempty"`
}

type UserService struct {
	ValidateUserID      func(string) (int64, error)
	ConnectDB           func() (*sql.DB, error)
	CheckUserNameExists func(*sql.DB, string) (bool, error)
	CreateUserFunc      func(*sql.DB, *User) (*User, error)
	UpdateUserFunc      func(*sql.DB, *User) (*User, error)
	DeleteUserFunc      func(*sql.DB, int64) error
	GetUserFunc         func(*sql.DB, int64) (*User, error)
	GetAllUsersFunc     func(*sql.DB) ([]User, error)
}

type Service interface {
	GetAll(c echo.Context) ([]User, error)
	GetByID(c echo.Context) (*User, error)
	Create(c echo.Context, u *User) (*User, error)
	Update(c echo.Context, u *User) (*User, error)
	DeleteByID(c echo.Context) error
}

var _ Service = (*UserService)(nil)

// Gets a single user by `user_id`
func (us *UserService) GetByID(c echo.Context) (*User, error) {
	userID := c.Param("user_id")

	id, err := us.ValidateUserID(userID)
	if err != nil {
		return nil, err
	}

	dbcon, err := us.ConnectDB()
	if err != nil {
		return nil, err
	}
	defer dbcon.Close()

	user, err := us.GetUserFunc(dbcon, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Gets ALL users without pagination
func (us *UserService) GetAll(c echo.Context) ([]User, error) {
	dbcon, err := us.ConnectDB()
	if err != nil {
		return nil, err
	}
	defer dbcon.Close()

	users, err := us.GetAllUsersFunc(dbcon)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Creates a new user based on JSON body
func (us *UserService) Create(c echo.Context, reqUser *User) (*User, error) {
	dbcon, err := us.ConnectDB()
	if err != nil {
		return nil, err
	}
	defer dbcon.Close()

	user, err := us.CreateUserFunc(dbcon, reqUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Updates a single user based on JSON body
func (us *UserService) Update(c echo.Context, reqUser *User) (*User, error) {
	dbcon, err := us.ConnectDB()
	if err != nil {
		return nil, err
	}
	defer dbcon.Close()

	user, err := us.UpdateUserFunc(dbcon, reqUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Delete user by `user_id`
func (us *UserService) DeleteByID(c echo.Context) error {
	userID := c.Param("user_id")

	id, err := us.ValidateUserID(userID)
	if err != nil {
		return err
	}

	dbcon, err := us.ConnectDB()
	if err != nil {
		return err
	}
	defer dbcon.Close()

	err = us.DeleteUserFunc(dbcon, id)
	if err != nil {
		return err
	}

	return nil
}
