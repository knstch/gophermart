package main

import (
	"bytes"
	"context"
	"database/sql"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/common"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/router"
	"github.com/knstch/gophermart/internal/app/storage/psql"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func loginGenerator(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

type testUser struct {
	login    string
	password string
}

var testUserOne = testUser{
	login:    loginGenerator(10),
	password: "12345",
}

var testUserTwo = testUser{
	login:    loginGenerator(10),
	password: "12345",
}

var testUserThree = testUser{
	login:    loginGenerator(10),
	password: "12345",
}

var orderNum = "5105105105105100"

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
				body:        `{"login": "` + testUserOne.login + `","password": "` + testUserOne.password + `"}`,
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
				body:        `{"login": "` + testUserOne.login + `","password": "` + testUserOne.password + `"}`,
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
				body:        `{"password": "` + testUserOne.password + `"}`,
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
				body:        `{"login": "` + testUserOne.login + `"}`,
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
				body:        `{"login": "` + testUserOne.login + `","password": ""}`,
			},
		},
		{
			name: "#8 sign up another user",
			want: want{
				statusCode:  200,
				contentType: "application/json; charset=utf-8",
				body:        `{"message":"Successfully registered"}`,
			},
			reqest: request{
				contentType: "application/json",
				body:        `{"login": "` + testUserTwo.login + `","password": "` + testUserTwo.password + `"}`,
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
				body:        `{"login": "` + testUserOne.login + `","password": "` + testUserOne.password + `"}`,
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
				body:        `{"213ssd": "` + testUserOne.login + `","passwordasd": "` + testUserOne.password + `"}`,
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
				body:        `{"login": "` + testUserOne.login + `14","password": "` + testUserOne.password + `"}`,
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
		user        testUser
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
				body:        orderNum,
				user:        testUserOne,
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
				body:        orderNum,
				user:        testUserOne,
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
				user:        testUserOne,
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
				user:        testUserTwo,
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
				user:        testUserOne,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCookieRes := httptest.NewRecorder()
			if tt.name != "#6 upload order without cookie" {
				getCookieReqBody := `{"login": "` + tt.reqest.user.login + `","password": "` + tt.reqest.user.password + `"}`
				getCookieReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", bytes.NewBuffer([]byte(getCookieReqBody)))
				getCookieReq.Header.Set("Content-Type", tt.reqest.contentType)
				router.ServeHTTP(getCookieRes, getCookieReq)
			}

			cookies := getCookieRes.Result().Cookies()

			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/orders/", bytes.NewBuffer([]byte(tt.reqest.body)))
			req.Header.Set("Content-Type", tt.reqest.contentType)
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}

func TestGetOrders(t *testing.T) {
	db, err := sql.Open("pgx", config.ReadyConfig.Database)
	if err != nil {
		logger.ErrorLogger("Can't open connection: ", err)
	}
	storage := psql.NewPsqlStorage(db)

	h := handler.NewHandler(storage)

	router := router.RequestsRouter(h)

	var orderTest common.Order
	ctx := context.Background()
	dbBun := bun.NewDB(db, pgdialect.New())
	err = dbBun.NewSelect().
		Model(&orderTest).
		Where(`"order" = ?`, orderNum).
		Scan(ctx)
	if err != nil {
		logger.ErrorLogger("Error making request to DB", err)
	}

	type want struct {
		statusCode  int
		contentType string
		body        string
	}

	type request struct {
		user testUser
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
						"number": "` + orderTest.Order + `",
						"status": "` + orderTest.Status + `",
						"uploaded_at": "` + orderTest.UploadedAt + `",
						"BonusesWithdrawn": null,
						"accrual": null
					}
				]`,
			},
			reqest: request{
				user: testUserOne,
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
				user: testUserThree,
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
				user: testUserThree,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCookieRes := httptest.NewRecorder()

			switch tt.name {
			case "#2 if user don't have orders":
				getCookieReqBody := `{"login": "` + tt.reqest.user.login + `","password": "` + tt.reqest.user.password + `"}`
				getCookieReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/register", bytes.NewBuffer([]byte(getCookieReqBody)))
				getCookieReq.Header.Set("Content-Type", "application/json; charset=utf-8")
				router.ServeHTTP(getCookieRes, getCookieReq)

			case "#3 user without cookie":

			default:
				getCookieReqBody := `{"login": "` + tt.reqest.user.login + `","password": "` + tt.reqest.user.password + `"}`
				getCookieReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", bytes.NewBuffer([]byte(getCookieReqBody)))
				getCookieReq.Header.Set("Content-Type", "application/json; charset=utf-8")
				router.ServeHTTP(getCookieRes, getCookieReq)
			}

			cookies := getCookieRes.Result().Cookies()

			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/user/orders/", nil)
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
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
