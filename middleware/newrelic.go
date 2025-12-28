package middleware

import (
	"fmt"
	"gaming-leaderboard/constants"
	"gaming-leaderboard/pkg/env"
	"net/http"
	"time"

	onewrelic "gaming-leaderboard/pkg/newrelic"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicMiddleware creates a production-grade New Relic monitoring middleware
func NewRelicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		txn := onewrelic.NRApp.StartTransaction(c.Request.Method + " " + c.FullPath())
		defer txn.End()

		// Set HTTP details
		txn.SetWebRequestHTTP(c.Request)
		txn.SetWebResponse(c.Writer)

		// Add custom attributes for better insights
		txn.AddAttribute("http.method", c.Request.Method)
		txn.AddAttribute("http.url", c.FullPath())
		txn.AddAttribute("http.host", c.Request.Host)
		txn.AddAttribute("http.useragent", c.Request.UserAgent())
		txn.AddAttribute(constants.RequestID, env.GetRequestID(c))

		if userID := c.GetString(constants.UserID); userID != "" {
			txn.AddAttribute(constants.UserID, userID)
		}

		c.Request = c.Request.WithContext(
			newrelic.NewContext(c.Request.Context(), txn),
		)

		// Record timing
		startTime := time.Now()
		statusCode := http.StatusOK

		c.Next()

		// Capture actual status code
		statusCode = c.Writer.Status()
		duration := time.Since(startTime).Milliseconds()

		// Add response attributes
		txn.AddAttribute("http.status_code", statusCode)
		txn.AddAttribute("http.response_time_ms", duration)

		// Add custom metrics for slow requests (> 1000ms)
		if duration > 1000 {
			txn.AddAttribute("slow_request", true)
			txn.NoticeError(fmt.Errorf("slow request: %dms", duration))
		}

		// Handle errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				txn.NoticeError(err.Err)
			}
			txn.AddAttribute("has_errors", true)
			txn.AddAttribute("error_count", len(c.Errors))
		}
	}
}
