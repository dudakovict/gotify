package mid

import (
	"fmt"
	"strings"

	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/gofiber/fiber/v2"
)

// Authenticate authenticates user's payload and stores it in the context.
func Authenticate(maker maker.Maker, accessibleRoles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(authHeaderKey)
		if len(authHeader) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is not provided",
			})
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		authType := strings.ToLower(fields[0])
		if authType != authTypeBearer {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("unsupported authorization type %s", authType),
			})
		}

		accessToken := fields[1]
		payload, err := maker.VerifyToken(accessToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if !hasPermission(payload.Roles, accessibleRoles) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		c.Locals(AuthPayloadKey, payload)
		return c.Next()
	}
}

func hasPermission(userRoles []string, accessibleRoles []string) bool {
	for _, role := range accessibleRoles {
		roleFound := false
		for _, userRole := range userRoles {
			if role == userRole {
				roleFound = true
				break
			}
		}

		if !roleFound {
			return false
		}
	}

	return true
}
