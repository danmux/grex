package app

import (
	"testing"
)

func Test_InitGrex(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	InitGrex("/what/path", "my.eg.uri:8888", "2000", 20, 1)
}
