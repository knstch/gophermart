package gophermarterrors

import "errors"

// An error indicating that an order is loaded by another user.
var ErrAlreadyLoadedOrder = errors.New("order is loaded by another user")

// An error indicating that an order is loaded by user.
var ErrYouAlreadyLoadedOrder = errors.New("order is loaded by you")

// An error indicating that order number is not passed by Lunh algorythm.
var ErrWrongOrderNum = errors.New("wrong order number")

// An error indication that a users is not authenticated.
var ErrAuth = errors.New("you are not authenticated")

// An error indictating that a user has not enough balance.
var ErrNotEnoughBalance = errors.New("not enough balance")

// An error indiating that no rows were found.
var ErrNoRows = errors.New("no rows were found")

// Переместить ошибки по своим местам