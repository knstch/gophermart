// Package statuslogger provides a middleware to display requests and responses.
package statuslogger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Middleware for requests writing and logging URI, method, duration. Inside of the function
// we implement a new "logFn" function
// accepting request and response. In this function we swap a standard response interface
// to a modified to write the statuscode.
func WithLogger() gin.HandlerFunc {
	var logger, err = zap.NewDevelopment()
	var sugar = *logger.Sugar()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		statusCode := c.Writer.Status()
		duration := time.Since(start)

		sugar.Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"duration", duration,
			"status code", statusCode,
		)
	}
}
