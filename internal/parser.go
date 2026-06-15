package internal

import "fmt"

func SetPosition(name string) int {
	if name == "encore" {
		return 999
	}
	// extract the number from "set_1", "set_2", "set_3"
	var n int
	fmt.Sscanf(name, "set_%d", &n)
	return n
}
