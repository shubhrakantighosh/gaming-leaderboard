package middleware

import (
	"bytes"
	"io/ioutil"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}

		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Log Request
		log.Println("====== Incoming Request ======")
		log.Printf("Method: %s | URL: %s", c.Request.Method, c.Request.URL.String())
		log.Printf("Query Params: %v", c.Request.URL.Query())
		log.Printf("Body: %s", string(bodyBytes))

		// Capture response
		rw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = rw

		c.Next()

		// Log Response
		duration := time.Since(start)
		log.Println("====== Response ======")
		log.Printf("Status: %d | Duration: %v", c.Writer.Status(), duration)
		log.Printf("Response Body: %s", rw.body.String())
	}
}
