package user_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/steveperjesi/integra-demo/user"
)

var _ = Describe("UserService", func() {
	var (
		e      *echo.Echo
		us     *user.UserService
		mockDB *sql.DB
	)

	BeforeEach(func() {
		e = echo.New()
		var err error
		mockDB, _, err = sqlmock.New()
		Expect(err).To(BeNil())

		us = &user.UserService{
			ConnectDB: func() (*sql.DB, error) { return mockDB, nil },
			ValidateUserID: func(s string) (int64, error) {
				if s == "bad" {
					return 0, errors.New("invalid ID")
				}
				return 123, nil
			},
			GetUserFunc: func(db *sql.DB, id int64) (*user.User, error) {
				return &user.User{ID: id, UserName: "testuser"}, nil
			},
			GetAllUsersFunc: func(db *sql.DB) ([]user.User, error) {
				return []user.User{{ID: 1, UserName: "alice"}}, nil
			},
			CreateUserFunc: func(db *sql.DB, u *user.User) (*user.User, error) {
				u.ID = 101
				return u, nil
			},
			UpdateUserFunc: func(db *sql.DB, u *user.User) (*user.User, error) {
				u.UserName = "updated"
				return u, nil
			},
			DeleteUserFunc: func(db *sql.DB, id int64) error {
				return nil
			},
		}
	})

	AfterEach(func() {
		mockDB.Close()
	})

	It("GetByID returns a user", func() {
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/users/123", nil), httptest.NewRecorder())
		c.SetParamNames("user_id")
		c.SetParamValues("123")

		u, err := us.GetByID(c)
		Expect(err).To(BeNil())
		Expect(u.UserName).To(Equal("testuser"))
	})

	It("GetAll returns users", func() {
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/users", nil), httptest.NewRecorder())
		u, err := us.GetAll(c)
		Expect(err).To(BeNil())
		Expect(u).To(HaveLen(1))
		Expect(u[0].UserName).To(Equal("alice"))
	})

	It("Create creates a new user", func() {
		req := &user.User{UserName: "newuser"}
		c := e.NewContext(nil, nil)
		u, err := us.Create(c, req)
		Expect(err).To(BeNil())
		Expect(u.ID).To(Equal(int64(101)))
	})

	It("Update updates a user", func() {
		req := &user.User{ID: 5, UserName: "old"}
		c := e.NewContext(nil, nil)
		u, err := us.Update(c, req)
		Expect(err).To(BeNil())
		Expect(u.UserName).To(Equal("updated"))
	})

	It("DeleteByID deletes a user", func() {
		c := e.NewContext(nil, nil)
		c.SetParamNames("user_id")
		c.SetParamValues("123")

		err := us.DeleteByID(c)
		Expect(err).To(BeNil())
	})

	It("GetByID returns error on bad ID", func() {
		c := e.NewContext(nil, nil)
		c.SetParamNames("user_id")
		c.SetParamValues("bad")

		_, err := us.GetByID(c)
		Expect(err).To(MatchError("invalid ID"))
	})
})
