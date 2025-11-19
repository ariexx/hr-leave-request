package middleware

import (
	"hr-leave-request/config"
	"hr-leave-request/dtos"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(cfg *config.ApplicationConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token claims",
			})
		}

		// Store user info in context
		c.Locals("user_id", uint(claims["user_id"].(float64)))
		c.Locals("email", claims["email"].(string))
		c.Locals("role", claims["role"])

		return c.Next()
	}
}
