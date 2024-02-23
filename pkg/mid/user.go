package mid

import (
	"errors"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/gofiber/fiber/v2"
)

// GetUser stores a user in the context.
func GetUser(usrCore *user.Core) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload, ok := c.Locals(AuthPayloadKey).(*maker.Payload)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get user payload",
			})
		}

		usr, err := usrCore.QueryByID(payload.UserID)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		c.Locals(UserKey, usr)
		return c.Next()
	}
}
