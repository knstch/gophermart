package main

import (
	"bytes"
	"database/sql"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/router"
	"github.com/knstch/gophermart/internal/app/storage/psql"
	"github.com/stretchr/testify/assert"
)

func loginGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func TestSignUp(t *testing.T) {
	config.ParseConfig()
	db, err := sql.Open("pgx", config.ReadyConfig.Database)
	if err != nil {
		logger.ErrorLogger("Can't open connection: ", err)
	}
	storage := psql.NewPsqlStorage(db)

	h := handler.NewHandler(storage)

	router := router.RequestsRouter(h)

	type want struct {
		statusCode  int
		contentType string
		body        string
	}

	type request struct {
		contentType string
		url         string
		body        string
	}

	testLogin := loginGenerator(10)

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 sign up test",
			want: want{
				statusCode:  200,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Successfully registered"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"login": "` + testLogin + `","password": "password123"}`,
			},
		},
		{
			name: "#2 login is already taken",
			want: want{
				statusCode:  409,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Login is already taken"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"login": "` + testLogin + `","password": "password123"}`,
			},
		},
		{
			name: "#3 bad request, no login",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"password": "password123"}`,
			},
		},
		{
			name: "#4 bad request, no password",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"login": "testuser1"}`,
			},
		},
		{
			name: "#5 bad request, wrong data",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"asd": "asd"}`,
			},
		},
		{
			name: "#6 bad request, empty login",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error": "Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"login": "","password": "password123"}`,
			},
		},
		{
			name: "#7 bad request, empty password",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				url:         "http://localhost:8080/api/user/register",
				body:        `{"login": "testuser1","password": ""}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.reqest.url, bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}
