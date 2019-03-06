package patch

import (
	"fmt"
	"testing"
)

func TestGetPatch(t *testing.T) {
	p := GetPatch(0, 0, "mesh", "eu-west-1", "vn", "80")
	fmt.Println("results")
	fmt.Println(string(p))
}
