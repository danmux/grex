package app

import (
	"testing"
)

func Test_CacheMissing(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	// make sure the cache isnt allocated
	seshCache = nil

	InitGrex("", "my.eg.uri", "8888", 10, 0)
	sesh1 := NewSesh()
	err := PutSeshInCache(sesh1)
	if err == nil {
		t.Error("did not ger an error from an empty cache")
	}
	t.Log(err)
}

func Test_SeshCache(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	InitGrex("", "my.eg.uri", "8888", 10, 1)

	sesh1 := NewSesh()
	sesh2 := NewSesh()

	t.Log(sesh1.Key)
	t.Log(sesh2.Key)

	// always different keys
	if sesh1.Key == sesh2.Key {
		t.Error("sesh keys the same")
	}

	// record the keys
	sKey1 := sesh1.Key
	sKey2 := sesh2.Key

	// set up two usernames
	u1 := "user-1-session"
	u2 := "user-2-session"

	sesh1.Uid = u1
	sesh2.Uid = u2

	// store them both
	PutSeshInCache(sesh1)
	PutSeshInCache(sesh2)

	testSesh1, in1, err := GetSeshFromCache(sKey1)

	if err != nil {
		t.Error(err)
	}

	if !in1 {
		t.Error("cache missing key 1")
	}

	if testSesh1.Uid != u1 {
		t.Error("sesh 1 not retrieved correctly")
	}

	testSesh2, in2, _ := GetSeshFromCache(sKey2)

	if !in2 {
		t.Error("cache missing key 2")
	}

	if testSesh2.Uid != u2 {
		t.Error("sesh 2 not retrieved correctly")
	}

	// now test for bad hits
	testSeshMissing, errorIn, _ := GetSeshFromCache("missingkey")

	if errorIn {
		t.Error("missing key odly found in cache")
	}

	if testSeshMissing != nil {
		t.Error("missing key odly found in cache")
	}

	testSesh2.Uid = "ts2"

	if sesh2.Uid == "ts2" {
		t.Error("some sesh2 reference hiccup " + sesh2.Uid)
	}
}
