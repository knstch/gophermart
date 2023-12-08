package validitycheck

import "strconv"

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
