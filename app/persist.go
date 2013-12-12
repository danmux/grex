//persist - write something to persistant storage
package app

import (
	"io/ioutil"
	"log"
	"os"
)

func getRootPath(request *BlobArg) string {
	return request.Key[:2] + "/" + request.Key
}

func getPath(request *BlobArg) string {
	return getRootPath(request) + "/" + request.SubKey + ".gob"
}

func GetData(request *BlobArg, response *BlobArg) error {

	log.Println("GET " + getPath(request))
	n, err := ioutil.ReadFile(getPath(request))
	if err != nil {
		return err
	}
	response.Message = "good"
	response.Payload = n
	response.Key = request.Key
	response.SubKey = request.SubKey

	log.Println("GET-done")
	return nil
}

func PostData(request *BlobArg, response *BlobArg) error {

	log.Println("write " + getPath(request))
	os.MkdirAll(getRootPath(request), 0750)
	err := ioutil.WriteFile(getPath(request), request.Payload, 0640)
	if err != nil {
		return err
	}

	// no payload for this response
	response.Message = "good"
	response.Key = request.Key
	response.SubKey = request.SubKey

	return nil
}
