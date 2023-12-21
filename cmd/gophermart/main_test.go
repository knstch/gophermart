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

func LoginGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

var testLogin = "aboba"
var testPassword = "12345"

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
		body        string
	}

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
				body:        `{"login": "` + testLogin + `","password": "` + testPassword + `"}`,
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
				body:        `{"login": "` + testLogin + `","password": "` + testPassword + `"}`,
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
				body:        `{"password": "` + testPassword + `"}`,
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
				body:        `{"login": "` + testLogin + `"}`,
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
				body:        `{"login": "` + testLogin + `","password": ""}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/register", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestAuth(t *testing.T) {
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
		body        string
	}

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 sign in test",
			want: want{
				statusCode:  200,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Successfully signed in"}`,
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"login": "` + testLogin + `","password": "` + testPassword + `"}`,
			},
		},
		{
			name: "#2 bad request, wrong request",
			want: want{
				statusCode:  400,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong request"}`,
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"213ssd": "` + testLogin + `","passwordasd": "` + testPassword + `"}`,
			},
		},
		{
			name: "#3 unauthorized, wrong credentials",
			want: want{
				statusCode:  401,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong email or password"}`,
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"login": "` + testLogin + `14","password": "` + testPassword + `"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestUploadOrder(t *testing.T) {
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
		body        string
		cookie      http.Cookie
	}

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 upload good order test",
			want: want{
				statusCode:  202,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Successfully loaded order"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `5105105105105100`,
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZSJ9.oO_0Lz-4YxuNZFhX5Od2Do1Kr-3srOnN5cGPuyKzZ3Q",
					Path:  "/",
				},
			},
		},
		{
			name: "#2 reload good order",
			want: want{
				statusCode:  200,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Order is already loaded"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `5105105105105100`,
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZSJ9.oO_0Lz-4YxuNZFhX5Od2Do1Kr-3srOnN5cGPuyKzZ3Q",
					Path:  "/",
				},
			},
		},
		{
			name: "#3 wrong order number",
			want: want{
				statusCode:  422,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Wrong order number"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `12345`,
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZSJ9.oO_0Lz-4YxuNZFhX5Od2Do1Kr-3srOnN5cGPuyKzZ3Q",
					Path:  "/",
				},
			},
		},
		{
			name: "#4 order load by another user",
			want: want{
				statusCode:  202,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Successfully loaded order"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `30569309025904`,
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZTIifQ.s0v16E6uFXT6iwodOXJhenKvof2_dvg6qOwbFr2IsO8",
					Path:  "/",
				},
			},
		},
		{
			name: "#5 upload order uploaded by another user",
			want: want{
				statusCode:  409,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"Order is already loaded by another user"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `30569309025904`,
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZSJ9.oO_0Lz-4YxuNZFhX5Od2Do1Kr-3srOnN5cGPuyKzZ3Q",
					Path:  "/",
				},
			},
		},
		{
			name: "#6 upload order without cookie",
			want: want{
				statusCode:  401,
				contentType: "application/json; charset=utf-8",
				body:        `{"error":"You are not authenticated"}`,
			},
			reqest: request{
				contentType: "text/plain; charset=utf-8",
				body:        `30569309025904`,
				cookie: http.Cookie{
					Name:  "1",
					Value: "2",
					Path:  "/",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/orders/", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			req.AddCookie(&tt.reqest.cookie)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestGetOrders(t *testing.T) {
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
		cookie http.Cookie
	}

	tests := []struct {
		name   string
		want   want
		reqest request
	}{
		{
			name: "#1 get orders if user have",
			want: want{
				statusCode:  200,
				contentType: "application/json; charset=utf-8",
				body: `[
					{
						"number": "5105105105105100",
						"status": "NEW",
						"uploaded_at": "2023-12-21T15:09:05Z",
						"BonusesWithdrawn": null,
						"accrual": null
					}
				]`,
			},
			reqest: request{
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZSJ9.oO_0Lz-4YxuNZFhX5Od2Do1Kr-3srOnN5cGPuyKzZ3Q",
					Path:  "/",
				},
			},
		},
		{
			name: "#2 if user don't have orders",
			want: want{
				statusCode:  204,
				contentType: "application/json; charset=utf-8",
				body:        "",
			},
			reqest: request{
				cookie: http.Cookie{
					Name:  "Auth",
					Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImFiZTMifQ.D0GQrUu9iIPfjP0_qNW7W7b_dHJ2dL7gzE0kMcAtGD8",
					Path:  "/",
				},
			},
		},
		{
			name: "#3 user without cookie",
			want: want{
				statusCode:  401,
				contentType: "application/json; charset=utf-8",
				body:        `{"error": "You are not authenticated"}`,
			},
			reqest: request{
				cookie: http.Cookie{
					Name:  "1",
					Value: "2",
					Path:  "/",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/user/orders/", nil)
			req.AddCookie(&tt.reqest.cookie)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			if tt.want.body != "" {
				assert.JSONEq(t, tt.want.body, rr.Body.String())
			}
		})
	}
}
