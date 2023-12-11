package statuslogger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData saves response status code.
type responseData struct {
	status int
}

// loggingResponse contains responseData to get data of a request
// and implements http.ResponseWriter interface.
type loggingResponse struct {
	http.ResponseWriter
	responseData *responseData
}

// Modification of the WriteHeader intefrace saving status code to a variable.
func (r *loggingResponse) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Middleware for requests writing and logging URI, method, duration. It accepts handler
// and returns handler. Inside of the function we implement a new "logFn" function
// accepting request and response. In this function we swap a standard response interface
// to a modified to write the statuscode.
func WithLogger(h http.Handler) http.Handler {
	var logger, err = zap.NewDevelopment()
	var sugar = *logger.Sugar()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logFn := func(res http.ResponseWriter, req *http.Request) {
		responseData := &responseData{
			status: 0,
		}
		loggingRes := loggingResponse{
			ResponseWriter: res,
			responseData:   responseData,
		}

		start := time.Now()

		uri := req.RequestURI

		method := req.Method

		h.ServeHTTP(&loggingRes, req)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status code", responseData.status,
		)
	}
	return http.HandlerFunc(logFn)
}
