package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nikola43/aureo-vpn/pkg/auth"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(tokenService *auth.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := parts[1]

		// Verify token
		claims, err := tokenService.VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("username", claims.Username)
		c.Locals("is_admin", claims.IsAdmin)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// AdminOnlyMiddleware ensures only admin users can access the route
func AdminOnlyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		isAdmin, ok := c.Locals("is_admin").(bool)
		if !ok || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin access required",
			})
		}

		return c.Next()
	}
}
