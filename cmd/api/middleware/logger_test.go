package middleware

import "testing"

func TestRedactIP(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"ipv4", "192.168.0.1", "192.168.0.x"},
		{"ipv4 with port", "10.0.0.100:8080", "10.0.0.x"},
		{"ipv6", "2345:0425:2CA1::0567:5673:23b5", "2345:425:2ca1:0:x:x:x:x"},
	}

	for _, tt := range tests {
		got := redactIP(tt.in)

		if got != tt.want {
			t.Errorf("got %q; want %q", got, tt.want)
		}
	}
}
