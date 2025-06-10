package user

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type MockUserService struct {
	GetAllFunc     func(c echo.Context) ([]User, error)
	GetByIDFunc    func(c echo.Context) (*User, error)
	CreateFunc     func(c echo.Context, u *User) (*User, error)
	UpdateFunc     func(c echo.Context, u *User) (*User, error)
	DeleteByIDFunc func(c echo.Context) error
}

func (m *MockUserService) GetAll(c echo.Context) ([]User, error) {
	if m.GetAllFunc == nil {
		return nil, errors.New("GetAllFunc not implemented")
	}
	return m.GetAllFunc(c)
}

func (m *MockUserService) GetByID(c echo.Context) (*User, error) {
	if m.GetByIDFunc == nil {
		return nil, errors.New("GetByIDFunc not implemented")
	}
	return m.GetByIDFunc(c)
}

func (m *MockUserService) Create(c echo.Context, u *User) (*User, error) {
	if m.CreateFunc == nil {
		return nil, errors.New("CreateFunc not implemented")
	}
	return m.CreateFunc(c, u)
}

func (m *MockUserService) Update(c echo.Context, u *User) (*User, error) {
	if m.UpdateFunc == nil {
		return nil, errors.New("UpdateFunc not implemented")
	}
	return m.UpdateFunc(c, u)
}

func (m *MockUserService) DeleteByID(c echo.Context) error {
	if m.DeleteByIDFunc == nil {
		return errors.New("DeleteByIDFunc not implemented")
	}
	return m.DeleteByIDFunc(c)
}
