// Package validity check provides a function that checks correctness of an order number.
package validitycheck

import (
	"errors"
	"strconv"
)

// LuhnAlgorithm checks an order number by luhn algorithm
// and returns true if the number is correct and false if it's wrong.
func LuhnAlgorithm(orderNumber string) bool {
	runes := []rune(orderNumber)
	sum := 0
	for i := len(runes) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(runes[i]))
		if (len(runes)-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

// An error indicating that order number is not passed by Lunh algorythm.
var ErrWrongOrderNum = errors.New("wrong order number")
