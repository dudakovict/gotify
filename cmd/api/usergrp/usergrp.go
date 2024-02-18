// Package usergrp maintains the group of handlers for user access.
package usergrp

import (
	"errors"
	"slices"
	"time"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/dudakovict/gotify/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type handlers struct {
	user *user.Core
}

func new(user *user.Core) *handlers {
	return &handlers{
		user: user,
	}
}

type createUserRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Roles    []string `json:"roles" validate:"required"`
	Password string   `json:"password" validate:"required,min=6"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newUserResponse(usr user.User) userResponse {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return userResponse{
		ID:        usr.ID,
		Email:     usr.Email,
		Roles:     roles,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
	}
}

// @Summary Create a new user
// @Description Creates a new user with the provided email, roles, and password.
// @Tags User
// @Accept json
// @Produce json
// @Param body body createUserRequest true "Create User Request Body"
// @Success 201 {object} userResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [post]
func (h *handlers) create(c *fiber.Ctx) error {
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	roles, err := parseRoles(req.Roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	nu := user.NewUser{
		Email:    req.Email,
		Roles:    roles,
		Password: req.Password,
	}

	newUsr, err := h.user.Create(nu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return c.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusCreated).JSON(newUserResponse(newUsr))
}

type updateUserRequest struct {
	Email    *string  `json:"email" validate:"omitempty,email"`
	Roles    []string `json:"roles"`
	Password *string  `json:"password"`
	Verified *bool    `json:"verified"`
}

// @Summary Update a user
// @Description Updates a user with the provided email, roles, password, and verification status.
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body updateUserRequest true "Update User Request Body"
// @Success 200 {object} userResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [put]
func (h *handlers) update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	var req updateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != id {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	var roles []user.Role
	if req.Roles != nil {
		roles, err = parseRoles(req.Roles)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		}
	}

	queryUsr, err := h.user.QueryByID(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	uu := user.UpdateUser{
		Email:    req.Email,
		Roles:    roles,
		Password: req.Password,
		Verified: req.Verified,
	}

	updUsr, err := h.user.Update(queryUsr, uu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return c.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusOK).JSON(newUserResponse(updUsr))
}

// @Summary Delete a user
// @Description Deletes a user by its ID.
// @Tags User
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [delete]
func (h *handlers) delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != id {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	queryUsr, err := h.user.QueryByID(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	if err := h.user.Delete(queryUsr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Query users
// @Description Queries users based on page number and rows per page.
// @Tags User
// @Produce json
// @Param page query integer true "Page Number"
// @Param rows query integer true "Rows Per Page"
// @Success 200 {array} userResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [get]
func (h *handlers) query(c *fiber.Ctx) error {
	page, err := util.Parse(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	users, err := h.user.Query(page.Number, page.RowsPerPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	usrsResponse := make([]userResponse, len(users))
	for i, user := range users {
		usrsResponse[i] = newUserResponse(user)
	}

	return c.Status(fiber.StatusOK).JSON(usrsResponse)
}

// @Summary Query a user by ID
// @Description Queries a user by its ID.
// @Tags User
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} userResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [get]
func (h *handlers) queryByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != id {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	queryUsr, err := h.user.QueryByID(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusOK).JSON(newUserResponse(queryUsr))
}
