package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/steveperjesi/integra-demo/user"
)

// @Summary      Get all users
// @Description  Retrieves all user information
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200 {object} []user.User
// @Failure      500 {object} ErrorResponse
// @Router       /users [get]
func GetAllUsers(service user.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		users, err := service.GetAll(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusOK, users)
	}
}

// @Summary      Get a user by ID
// @Description  Retrieves user information by user_id
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      200 {object} user.User
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /users/{user_id} [get]
func GetUserByID(service user.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := service.GetByID(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusOK, user)
	}
}

// @Summary      Create a new user
// @Description  Creates a new user from a user request body
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user body user.User true "User data"
// @Success      201 {object} user.User
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /users [post]
func CreateUser(service user.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var userRequest user.User
		if err := c.Bind(&userRequest); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		if err := userRequest.ValidateNewUserRequest(); err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		newUser, err := service.Create(c, &userRequest)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusCreated, newUser)
	}
}

// @Summary      Update an existing user
// @Description  Updates an existing user based on the given body
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user body user.User true "Updated user data"
// @Success      200 {object} user.User
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /users [put]
func UpdateUser(service user.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var userRequest user.User
		if err := c.Bind(&userRequest); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		updatedUser, err := service.Update(c, &userRequest)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusOK, updatedUser)
	}
}

// @Summary      Delete a user
// @Description  Deletes a user by user_id
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user_id path string true "User ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /users/{user_id} [delete]
func DeleteUser(service user.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := service.DeleteByID(c); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return c.NoContent(http.StatusNoContent)
	}
}
