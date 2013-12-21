package app

import (
	"testing"
)

func Test_GetVersion(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	InitGrex("", "my.eg.uri", "8888", 10, 0)

	resetBucketVersion("danm")

	bv := getBucketVersion("danm")

	if bv.getVersion("accounts") != 0 {
		t.Error("getVersion Failed")
	}

	bv.incVersion("accounts")

	t.Log(bv.getVersion("accounts"))
	if bv.getVersion("accounts") != 1 {
		t.Error("incVersion Failed")
	}
}

func Test_GetVersionFromDisk(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T
	InitGrex("", "my.eg.uri", "8888", 10, 0)
	resetBucketVersion("danm")

	bv := getBucketVersion("danm")

	bv.incVersion("accounts")
	bv.incVersion("accounts")
	bv.incVersion("accounts")

	if bv.getVersion("accounts") != 3 {
		t.Error("did not inc the versions properly")
	}

	t.Log("putting")
	ch := putBucketVersion(bv)

	// wait for ersponse of actually perisiting to disk
	err := <-ch

	t.Log("put")
	if err != nil {
		t.Error(err)
	}

	clearVersionCache()

	nbv := getBucketVersion("danm")
	ver := nbv.getVersion("accounts")

	if ver != 3 {
		t.Error("did not persist the version to disk")
	}

}
