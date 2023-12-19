package validitycheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		request string
	}{
		{
			name:    "#1 valid data",
			want:    true,
			request: "5105105105105100",
		},
		{
			name:    "#2 valid data",
			want:    true,
			request: "30569309025904",
		},
		{
			name:    "#2 invalid data",
			want:    false,
			request: "12345",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderNumber := LuhnAlgorithm(tests[i].request)
			assert.Equal(t, tests[i].want, orderNumber)
		})
	}
}
