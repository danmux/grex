package app

import (
	"testing"
)

type TestObj struct {
	Who  string
	What string
	When int
}

func (t TestObj) Size() int {
	return 40
}

func Test_DataCacheMissing(t *testing.T) {
	// make sure the cache isnt allocated
	seshCache = nil

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 0)

	w := TestObj{
		"me",
		"fart",
		11,
	}

	PutItemInCache("danm", "testobj", w)
}

func Test_DataCache(t *testing.T) {
	InitGrex("../testdata", "my.eg.uri", "8888", 10, 1)

	w := TestObj{
		"meen",
		"farty",
		12,
	}

	PutItemInCache("danx", "testobject", w)

	y, in := GetItemFromCache("danx", "testobject")

	if !in {
		t.Error("cache missing key danm testobj")
	}

	if y.(TestObj).Who != "meen" {
		t.Error("cache returned some dodgy stuff")
	}
}
