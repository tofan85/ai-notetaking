package middleware

import (
	"runtime"
	"time"

	"ai-notetaking-be/internal/loggers"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func LoggingMiddleware(logger loggers.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		before := m.Alloc

		err := c.Next()
		traceID := uuid.New().String()
		c.Locals("trace_id", traceID)
		duration := time.Since(start).Milliseconds()
		runtime.ReadMemStats(&m)
		used := m.Alloc - before
		fields := map[string]interface{}{
			"trace_id":  traceID,
			"method":    c.Method(),
			"path":      c.Path(),
			"status":    c.Response().StatusCode(),
			"duration":  duration,
			"memory_kb": used / 1024,
		}

		if err != nil {
			fields["error"] = err.Error()
			logger.Error("HTTP Request FAILED", fields)
			return err
		}

		logger.Info("HTTP Request SUCCESS", fields)
		return nil
	}
}
