package statusLogger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
	}

	loggingResponse struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Модификация интерфейса WriteHeader, добавляем сохрание статус кода в переменную
func (r *loggingResponse) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Middlware обработчик для запросов, записывает URI, method, duration
func RequestsLogger(h http.Handler) http.Handler {
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
