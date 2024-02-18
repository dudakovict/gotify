// Package topicgrp maintains the group of handlers for topic access.
package topicgrp

import (
	"errors"
	"time"

	"github.com/dudakovict/gotify/internal/core/topic"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/dudakovict/gotify/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type handlers struct {
	topic *topic.Core
}

func new(topic *topic.Core) *handlers {
	return &handlers{
		topic: topic,
	}
}

type createTopicRequest struct {
	Name string `json:"name" validate:"required"`
}

type topicResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func newTopicResponse(tpc topic.Topic) topicResponse {
	return topicResponse{
		ID:        tpc.ID,
		Name:      tpc.Name,
		CreatedAt: tpc.CreatedAt,
	}
}

// @Summary Create a new topic
// @Description Creates a new topic with the provided name.
// @Tags Topic
// @Accept json
// @Produce json
// @Param body body createTopicRequest true "Create Topic Request Body"
// @Success 201 {object} topicResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /topics [post]
func (h *handlers) create(c *fiber.Ctx) error {
	var req createTopicRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	if err := validate.Check(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	nt := topic.NewTopic{
		Name: req.Name,
	}

	tpc, err := h.topic.Create(nt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusCreated).JSON(newTopicResponse(tpc))
}

// @Summary Query topics
// @Description Queries topics based on page number and rows per page.
// @Tags Topic
// @Produce json
// @Param page query integer true "Page Number"
// @Param rows query integer true "Rows Per Page"
// @Success 200 {array} topicResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /topics [get]
func (h *handlers) query(c *fiber.Ctx) error {
	page, err := util.Parse(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	tpcs, err := h.topic.Query(page.Number, page.RowsPerPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	tpcsResponse := make([]topicResponse, len(tpcs))
	for i, tpc := range tpcs {
		tpcsResponse[i] = newTopicResponse(tpc)
	}

	return c.Status(fiber.StatusOK).JSON(tpcsResponse)
}

// @Summary Query a topic by ID
// @Description Queries a topic by its ID.
// @Tags Topic
// @Produce json
// @Param id path string true "Topic ID"
// @Success 200 {object} topicResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /topics/{id} [get]
func (h *handlers) queryByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	}

	tpc, err := h.topic.QueryByID(id)
	if err != nil {
		if errors.Is(err, topic.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse(err))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	return c.Status(fiber.StatusOK).JSON(newTopicResponse(tpc))
}
