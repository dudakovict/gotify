// Package vrfgrp maintains the group of handlers for verification access.
package vrfgrp

import (
	"github.com/dudakovict/gotify/internal/core/verification"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type handlers struct {
	verification *verification.Core
}

func new(verification *verification.Core) *handlers {
	return &handlers{
		verification: verification,
	}
}

// @Summary Verify user
// @Description Verify user with the provided ID.
// @Tags Verification
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /verify [get]
func (h *handlers) verify(c *fiber.Ctx) error {
	id := c.Query("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	usr, err := h.verification.Verify(parsedID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	}

	rsp := struct {
		Verified bool `json:"verified"`
	}{
		Verified: usr.Verified,
	}

	return c.Status(fiber.StatusOK).JSON(rsp)
}
