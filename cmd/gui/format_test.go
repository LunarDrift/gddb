package main

import (
	"testing"
)

func TestFormatSetName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "set name with underscore", input: "set_1", want: "Set 1"},
		{name: "no underscore", input: "acoustic", want: "Acoustic"},
		{name: "extra whitespace", input: "   encore ", want: "Encore"},
		{name: "empty input", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSetName(tt.input)
			if got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}
