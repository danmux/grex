package app

//persist - write something to persistant storage

import (
	"io/ioutil"
	"log"
	"os"
)

func getRootPath(blob *BlobArg) string {
	return dataRoot + getFlockKey(blob.Key) + "/" + blob.Key
}

func getPath(blob *BlobArg) string {
	return getRootPath(blob) + "/" + blob.SubKey + ".gob"
}

// read in data from the file system
func getData(blob *BlobArg) error {

	log.Println("Debug - reading local file: " + getPath(blob))
	n, err := ioutil.ReadFile(getPath(blob))
	if err != nil {
		return err
	}
	blob.Message = "good"
	blob.Payload = n

	return nil
}

// write data to the file system
func putDataIncVersion(blob *BlobArg) error {
	bv := getBucketVersion(blob.Key)
	err := putData(blob)
	if err == nil {
		bv.incVersion(blob.SubKey)
		putBucketVersion(bv)
	}
	return err
}

func putData(blob *BlobArg) error {

	// invalidate this in the data cache
	invalidateItemInCache(blob.Key, blob.SubKey)

	log.Println("writing local file " + getPath(blob))

	os.MkdirAll(getRootPath(blob), 0750)
	err := ioutil.WriteFile(getPath(blob), blob.Payload, 0640)
	if err != nil {
		blob.Message = "failed - " + err.Error()
		return err
	}

	blob.Message = "good"

	return nil
}

func deleteData(blob *BlobArg) error {

	log.Println("removing local file " + getPath(blob))

	err := os.Remove(getPath(blob))
	if err != nil {
		blob.Message = "failed to remove file - " + err.Error()
		return err
	}

	blob.Message = "good"
	return nil
}
