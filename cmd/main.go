package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/steveperjesi/integra-demo/docs"
	"github.com/steveperjesi/integra-demo/internal/db"
	"github.com/steveperjesi/integra-demo/internal/handlers"
	"github.com/steveperjesi/integra-demo/user"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/joho/godotenv"
)

// @title:			Demo user API
// @version:		0.0.0
// @description:	Demo user API for creating, updating, and deleting users
// @host:           localhost:8080
// @BasePath:       /

var (
	defaultServerPort = "8080"
)

func newUserService() *user.UserService {
	return &user.UserService{
		ConnectDB:           db.Connect,
		ValidateUserID:      user.ValidateUserID,
		CheckUserNameExists: user.CheckUserNameExists,
		CreateUserFunc:      user.CreateUser,
		UpdateUserFunc:      user.UpdateUser,
		DeleteUserFunc:      user.DeleteUser,
		GetUserFunc:         user.GetUser,
		GetAllUsersFunc:     user.GetAllUsers,
	}
}

func StartServer() *echo.Echo {
	e := echo.New()
	userService := newUserService()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "PONG")
	})

	e.GET("/users", handlers.GetAllUsers(userService))
	e.GET("/users/:user_id", handlers.GetUserByID(userService))
	e.POST("/users", handlers.CreateUser(userService))
	e.PUT("/users", handlers.UpdateUser(userService))
	e.DELETE("/users/:user_id", handlers.DeleteUser(userService))

	e.Static("/swagger", "swagger-ui")
	e.Static("/docs", "docs")

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	e := StartServer()

	port := os.Getenv("DEMO_PORT")
	if port == "" {
		// Set default port if ENV is empty
		port = defaultServerPort
	}

	s := http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		Handler:     e,
		ReadTimeout: 10 * time.Second,
	}

	if err := e.StartServer(&s); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
