// Package subgrp maintains the group of handlers for subscription access.
package subgrp

import (
	"errors"
	"slices"
	"time"

	"github.com/dudakovict/gotify/internal/core/subscription"
	"github.com/dudakovict/gotify/internal/core/topic"
	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/dudakovict/gotify/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type handlers struct {
	subscription *subscription.Core
	topic        *topic.Core
}

func new(subscription *subscription.Core, topic *topic.Core) *handlers {
	return &handlers{
		subscription: subscription,
		topic:        topic,
	}
}

type createSubscriptionRequest struct {
	TopicID uuid.UUID `json:"topic_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
}

type subscriptionResponse struct {
	ID        uuid.UUID `json:"id"`
	TopicID   uuid.UUID `json:"topic_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func newSubscriptionResponse(sub subscription.Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:        sub.ID,
		TopicID:   sub.TopicID,
		UserID:    sub.UserID,
		CreatedAt: sub.CreatedAt,
	}
}

// @Summary Create a new subscription
// @Description Creates a new subscription with the provided topic ID and user ID.
// @Tags Subscription
// @Accept json
// @Produce json
// @Param body body createSubscriptionRequest true "Create Subscription Request Body"
// @Success 201 {object} subscriptionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions [post]
func (h *handlers) create(c *fiber.Ctx) error {
	var req createSubscriptionRequest
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

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != req.UserID {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if _, err := h.topic.QueryByID(req.TopicID); err != nil {
		if errors.Is(err, topic.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	ns := subscription.NewSubscription{
		TopicID: req.TopicID,
		UserID:  req.UserID,
	}

	sub, err := h.subscription.Create(ns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusCreated).JSON(newSubscriptionResponse(sub))
}

// @Summary Delete a subscription by ID
// @Description Deletes a subscription with the specified ID.
// @Tags Subscription
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/{id} [delete]
func (h *handlers) delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	sub, err := h.subscription.QueryByID(id)
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != sub.UserID {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if err := h.subscription.Delete(sub); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Query subscriptions
// @Description Queries subscriptions based on page number and rows per page.
// @Tags Subscription
// @Produce json
// @Param page query integer true "Page Number"
// @Param rows query integer true "Rows Per Page"
// @Success 200 {array} subscriptionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions [get]
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

	subs, err := h.subscription.Query(page.Number, page.RowsPerPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	subsResponse := make([]subscriptionResponse, len(subs))
	for i, sub := range subs {
		subsResponse[i] = newSubscriptionResponse(sub)
	}

	return c.Status(fiber.StatusOK).JSON(subsResponse)
}

// @Summary Query a subscription by ID
// @Description Queries a subscription by its ID.
// @Tags Subscription
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} subscriptionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/{id} [get]
func (h *handlers) queryByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	sub, err := h.subscription.QueryByID(id)
	if err != nil {
		if errors.Is(err, topic.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) && usr.ID != sub.UserID {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	return c.Status(fiber.StatusOK).JSON(newSubscriptionResponse(sub))
}
