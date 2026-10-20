package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func BasicAuth(username, password string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		encoded, hasPrefix := strings.CutPrefix(header, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		user, pass, ok := strings.Cut(string(decoded), ":")
		if !hasPrefix || err != nil || !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			c.Set("WWW-Authenticate", `Basic realm="employee-api"`)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}
