# Accrual System API Server

## Introduction

This API server interacts with an accrual system to provide functionality related to user balance, authentication, and order management. It utilizes PostgreSQL for data storage, Gin framework for routing and serving requests. The server includes endpoints for retrieving user balance, withdrawing bonuses, user registration and authentication, as well as uploading and retrieving orders. Additionally, the server incorporates a middleware to check user authentication via cookies and compress data. 

## Endpoints
### Balance
1. **GET** /user/balance: Retrieve user's balance, including withdrawn amount.
2. **POST** /user/balance/withdraw: Withdraw user's bonuses.

### Auth
1. **POST** /user/register: User registration and authentication.
2. **POST** /user/login: User authentication and setting an auth cookie.

### Order
1. **POST** /user/orders: Upload order to the server.
2. **GET** /user/orders: Retrieve user's orders.
3. **GET** /user/withdrawals: Retrieve orders with spent bonuses.

## Database Initialization
The server initializes a PostgreSQL database with the following tables:
### Users Table

**Users**
| Login. Type:varchar(255),unique. | Password. Type:varchar(255) | Balance. Type:float | Withdrawn. Type:float |
|----------------------------------|-----------------------------|---------------------|-----------------------|
|                                  |                             |                     |                       |

**Orders**
| Login. Type:varchar(255). | Order. Type:varchar(255),unique | Status. Type:varchar(255) | UploadedAt. Type:timestamp | 
|---------------------------|---------------------------------|---------------------------|----------------------------|

BonusesWithdrawn. Type:float. | Accrual. Type:float. |
------------------------------|----------------------|

## Project Structure
The project contains the following folders:
+ cmd
  + accrual - Contains binary file of accrual system.
    + accrual_windows_amd64.exe - accrual system count logic.
  + config - Contains config package.
    + config.go - configuration setup functions and structs.
  + gophermart - Contains main package.
    + main.go - main function.
+ docs - swagger documentation.
+ internal - contains dir app where is all logic.
  + app - contains all logic.
    + common - contains common package, it has functions and structs that can be used from different packages.
      + common_structs.go - contains common structs that can be used from any package.
      + common.go - contains common functions that can be used from any package.
    + cookie - contains cookie package that is used to interact with cookies.
      + cookie.go - contains functions to make JWT, set auth cookie, get login from JWT.
    + handler - contains handler package with all handlers.
      + handler_structs.go - contains structs that are used to handler package.
      + handler.go - contains all handlers
    + logger - contains logger package with loggin functionality.
      + logger.go - contains functions for info and error logging.
    + middleware - contains middlewares.
      + cookieLogin - contains middleware working with auth cookies.
        + cookie_login.go - contains middleware checking auth status, parsing login and passing it thru context.
    + router - contains router package used to routing requests.
        + router.go - contains router.
    + storage - contains psql package working with PostgreSQL.
        + init.go - initializes PostgreSQL tables.
        + psql_storage_structs.go - contains structs used in psql package.
        + psql_storage.go - contains functions interacting with PostgreSQL. 
    + validityCheck - contains validitycheck package
        + validity_check.go - contains function checking validity of order number.
