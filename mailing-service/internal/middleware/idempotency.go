package middleware

import (
	"github.com/gin-gonic/gin"
)

const IdempotencyKeyHeader = "Idempotency-Key"
const IdempotencyKeyContext = "idempotency_key"

func Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			key := c.GetHeader(IdempotencyKeyHeader)
			if key != "" {
				c.Set(IdempotencyKeyContext, key)
			}
		}
		c.Next()
	}
}
