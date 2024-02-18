// Package authgrp maintains the group of handlers for auth access.
package authgrp

import (
	"context"
	"errors"
	"time"

	"github.com/dudakovict/gotify/internal/core/session"
	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/dudakovict/gotify/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type handlers struct {
	user                 *user.Core
	session              *session.Core
	maker                maker.Maker
	taskDistributor      worker.TaskDistributor
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func new(user *user.Core, session *session.Core, maker maker.Maker, taskDistributor worker.TaskDistributor, accessTokenDuration time.Duration, refreshTokenDuration time.Duration) *handlers {
	return &handlers{
		user:                 user,
		session:              session,
		maker:                maker,
		taskDistributor:      taskDistributor,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

type registerReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
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

// @Summary Register a new user
// @Description Registers a new user with the provided email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body registerReq true "Register Request Body"
// @Success 201 {object} userResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func (h *handlers) register(c *fiber.Ctx) error {
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	nuTx := user.NewUserTx{
		NewUser: user.NewUser{
			Email:    req.Email,
			Roles:    []user.Role{user.RoleUser},
			Password: req.Password,
		},
		AfterCreate: func(usr user.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Email: usr.Email,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return h.taskDistributor.DistributeTaskSendVerifyEmail(context.Background(), taskPayload, opts...)
		},
	}

	usr, err := h.user.CreateTx(nuTx)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return c.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusCreated).JSON(newUserResponse(usr))
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// @Summary Login user
// @Description Logs in a user with the provided email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body loginRequest true "Login Request Body"
// @Success 200 {object} loginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /login [post]
func (h *handlers) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	usr, err := h.user.Authenticate(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		case errors.Is(err, user.ErrAuthenticationFailure):
			return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		}
	}

	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	accessToken, accessPayload, err := h.maker.CreateToken(usr.ID, roles, h.accessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	refreshToken, refreshPayload, err := h.maker.CreateToken(usr.ID, roles, h.refreshTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	session, err := h.session.Create(session.NewSession{
		UserID:       usr.ID,
		RefreshToken: refreshToken,
		UserAgent:    c.Get("User-Agent"),
		ClientIP:     c.IP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	rsp := loginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiresAt,
		User:                  newUserResponse(usr),
	}

	return c.Status(fiber.StatusOK).JSON(rsp)
}
