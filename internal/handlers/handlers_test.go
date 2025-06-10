package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/steveperjesi/integra-demo/user"
)

var _ = Describe("GetAllUsers Handler", func() {
	var (
		e           *echo.Echo
		mockService *user.MockUserService
		handler     echo.HandlerFunc
		rec         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()

		mockService = &user.MockUserService{
			GetAllFunc: func(c echo.Context) ([]user.User, error) {
				return []user.User{
					{ID: 1, UserName: "jdoe", FirstName: "John", LastName: "Doe"},
				}, nil
			},
		}

		handler = GetAllUsers(mockService)
	})

	It("returns 200 and a list of users", func() {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusOK))

		var users []user.User
		err = json.NewDecoder(rec.Body).Decode(&users)
		Expect(err).To(BeNil())
		Expect(users).To(HaveLen(1))
		Expect(users[0].UserName).To(Equal("jdoe"))
	})
})

var _ = Describe("GetUserByID Handler", func() {
	var (
		e           *echo.Echo
		mockService *user.MockUserService
		handler     echo.HandlerFunc
		rec         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()

		mockService = &user.MockUserService{
			GetByIDFunc: func(c echo.Context) (*user.User, error) {
				return &user.User{ID: 1, UserName: "jdoe"}, nil
			},
		}
	})

	JustBeforeEach(func() {
		handler = GetUserByID(mockService)
	})

	It("returns 200 and the user", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("1")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("returns 500 when user_id is invalid", func() {
		mockService.GetByIDFunc = func(c echo.Context) (*user.User, error) {
			return nil, fmt.Errorf("invalid user_id: must be an integer")
		}

		req := httptest.NewRequest(http.MethodGet, "/users/foo", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("foo")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
	})

	It("returns 500 when user not found", func() {
		mockService.GetByIDFunc = func(c echo.Context) (*user.User, error) {
			return nil, fmt.Errorf("user not found")
		}

		req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("99")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
	})
})

var _ = Describe("CreateUser Handler", func() {
	var (
		e           *echo.Echo
		mockService *user.MockUserService
		handler     echo.HandlerFunc
		rec         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()

		mockService = &user.MockUserService{
			CreateFunc: func(c echo.Context, u *user.User) (*user.User, error) {
				return &user.User{
					ID:         1,
					UserName:   u.UserName,
					FirstName:  u.FirstName,
					LastName:   u.LastName,
					Email:      u.Email,
					UserStatus: u.UserStatus,
				}, nil
			},
		}
	})

	JustBeforeEach(func() {
		handler = CreateUser(mockService)
	})

	It("returns 201 on success", func() {
		body := `{
			"user_name":"jdoe",
			"first_name":"john",
			"last_name":"doe",
			"email":"jdoe@test.com",
			"department":"test"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusCreated))
	})

	It("returns 400 on bad JSON", func() {
		mockService.CreateFunc = func(c echo.Context, u *user.User) (*user.User, error) {
			return nil, fmt.Errorf("code=400, message=Syntax error: offset=14, error=invalid character '}' looking for beginning of value, internal=invalid character '}' looking for beginning of value")
		}

		body := `{"user_name":}` // malformed JSON
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 500 on service error", func() {
		mockService.CreateFunc = func(c echo.Context, u *user.User) (*user.User, error) {
			return nil, fmt.Errorf("missing first_name")
		}

		body := `{"user_name":"jdoe"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
	})
})

var _ = Describe("UpdateUser Handler", func() {
	var (
		e           *echo.Echo
		mockService *user.MockUserService
		handler     echo.HandlerFunc
		rec         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()

		mockService = &user.MockUserService{
			UpdateFunc: func(c echo.Context, u *user.User) (*user.User, error) {
				u.ID = 1
				u.UserName = "jdoe"
				u.FirstName = "john"
				u.LastName = "doe"
				u.Email = "jdoe@newemail.com"
				return u, nil
			},
		}
	})

	JustBeforeEach(func() {
		handler = UpdateUser(mockService)
	})

	It("returns 200 on success", func() {
		body := `{
			"user_id":1,
			"email":"jdoe@newemail.com"}`
		req := httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("returns 400 on bad JSON", func() {
		mockService.UpdateFunc = func(c echo.Context, u *user.User) (*user.User, error) {
			return nil, fmt.Errorf("code=400, message=Syntax error: offset=12, error=invalid character '}' looking for beginning of value, internal=invalid character '}' looking for beginning of value")
		}

		body := `{"user_id":}` // malformed JSON
		req := httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 500 on invalid user", func() {
		mockService.UpdateFunc = func(c echo.Context, u *user.User) (*user.User, error) {
			return nil, fmt.Errorf("no rows updated")
		}

		body := `{"user_id":45,"user_status":"A"}`
		req := httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
	})
})

var _ = Describe("DeleteUser Handler", func() {
	var (
		e           *echo.Echo
		mockService *user.MockUserService
		handler     echo.HandlerFunc
		rec         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		e = echo.New()
		rec = httptest.NewRecorder()

		mockService = &user.MockUserService{
			DeleteByIDFunc: func(c echo.Context) error {
				return nil
			},
		}
	})

	JustBeforeEach(func() {
		handler = DeleteUser(mockService)
	})

	It("returns 204 on success", func() {
		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("1")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusNoContent))
	})

	It("returns 400 on invalid ID", func() {
		mockService.DeleteByIDFunc = func(c echo.Context) error {
			return fmt.Errorf("invalid user_id: must be an integer")
		}

		req := httptest.NewRequest(http.MethodDelete, "/users/foo", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("foo")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 500 on service error", func() {
		mockService.DeleteByIDFunc = func(c echo.Context) error {
			return fmt.Errorf("user not found")
		}

		req := httptest.NewRequest(http.MethodDelete, "/users/99", nil)
		c := e.NewContext(req, rec)
		c.SetParamNames("user_id")
		c.SetParamValues("99")

		err := handler(c)
		Expect(err).To(BeNil())
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})
})
