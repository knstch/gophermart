package gophermarterrors

import "errors"

// An error indicating that an order is loaded by another user
var ErrAlreadyLoadedOrder = errors.New("order is loaded by another user")

// An error indicating that an order is loaded by user
var ErrYouAlreadyLoadedOrder = errors.New("order is loaded by you")

// An error indicating that order number is not passed by Lunh algorythm
var ErrWrongOrderNum = errors.New("wrong order number")

// An error indication that a users is not authenticated
var ErrAuth = errors.New("you are not authenticated")

var ErrNotEnoughBalance = errors.New("not enough balance")

var ErrNoRows = errors.New("no rows were found")
