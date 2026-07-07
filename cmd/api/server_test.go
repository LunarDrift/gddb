package main

import "testing"

func TestFuzzyPattern(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Hello world", "%Hello%world%"},
		{"Althea", "%Althea%"},
		{"Dark Star > drums > space > Dark Star", "%Dark%Star%>%drums%>%space%>%Dark%Star%"},
		{"", ""},
	}

	for _, tt := range tests {
		got := fuzzyPattern(tt.in)
		if got != tt.want {
			t.Errorf("fuzzyPattern(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
