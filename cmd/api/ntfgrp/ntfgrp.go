// Package ntfgrp maintains the group of handlers for notification access.
package ntfgrp

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/dudakovict/gotify/internal/core/notification"
	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/dudakovict/gotify/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type handlers struct {
	notification    *notification.Core
	taskDistributor worker.TaskDistributor
}

func new(notification *notification.Core, taskDistributor worker.TaskDistributor) *handlers {
	return &handlers{
		notification:    notification,
		taskDistributor: taskDistributor,
	}
}

type createNotificationRequest struct {
	TopicID uuid.UUID `json:"topic_id" validate:"required"`
	Message string    `json:"message" validate:"required"`
}

type notificationResponse struct {
	ID        uuid.UUID `json:"id"`
	TopicID   uuid.UUID `json:"topic_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func newNotificationResponse(ntf notification.Notification) notificationResponse {
	return notificationResponse{
		ID:        ntf.ID,
		TopicID:   ntf.TopicID,
		Message:   ntf.Message,
		CreatedAt: ntf.CreatedAt,
	}
}

// @Summary Create a new notification
// @Description Creates a new notification with the provided topic ID and message.
// @Tags Notification
// @Accept json
// @Produce json
// @Param body body createNotificationRequest true "Create Notification Request Body"
// @Success 201 {object} notificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications [post]
func (h *handlers) create(c *fiber.Ctx) error {
	var req createNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	nnTx := notification.NewNotificationTx{
		NewNotification: notification.NewNotification{
			TopicID: req.TopicID,
			Message: req.Message,
		},
		AfterCreate: func(ntf notification.Notification) error {
			taskPayload := &worker.PayloadSendNotification{
				NotificationID: ntf.ID,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return h.taskDistributor.DistributeTaskSendNotification(context.Background(), taskPayload, opts...)
		},
	}

	ntf, err := h.notification.CreateTx(nnTx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusCreated).JSON(newNotificationResponse(ntf))
}

// @Summary Delete a notification by ID
// @Description Deletes a notification with the specified ID.
// @Tags Notification
// @Produce json
// @Param id path string true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id} [delete]
func (h *handlers) delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	ntf, err := h.notification.QueryByID(id)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	usr, ok := c.Locals(mid.UserKey).(user.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if !slices.Contains(usr.Roles, user.RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(errorResponse(user.ErrAuthorizationFailure))
	}

	if err := h.notification.Delete(ntf); err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Query notifications
// @Description Queries notifications based on page number, rows per page, and topic ID.
// @Tags Notification
// @Produce json
// @Param page query integer true "Page Number"
// @Param rows query integer true "Rows Per Page"
// @Param topic_id query string true "Topic ID"
// @Success 200 {array} notificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications [get]
func (h *handlers) query(c *fiber.Ctx) error {
	page, err := util.Parse(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	topicID, err := uuid.Parse(c.Query("topic_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	ntfs, err := h.notification.Query(page.Number, page.RowsPerPage, topicID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	ntfsResponse := make([]notificationResponse, len(ntfs))
	for i, ntf := range ntfs {
		ntfsResponse[i] = newNotificationResponse(ntf)
	}

	return c.Status(fiber.StatusOK).JSON(ntfsResponse)
}

// @Summary Query a notification by ID
// @Description Queries a notification by its ID.
// @Tags Notification
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} notificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id} [get]
func (h *handlers) queryByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	ntf, err := h.notification.QueryByID(id)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusOK).JSON(newNotificationResponse(ntf))
}
