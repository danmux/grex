package app

import (
	"log"
)

type BlobArg struct {
	Key     string
	SubKey  string
	Message string
	Payload []byte
}

func GetBytes(key string, subkey string) (*BlobArg, error) {
	log.Println("GET")
	conn, err := GetConnection(":8008")
	if err != nil {
		return nil, err
	}

	args := &BlobArg{
		Key:    key,
		SubKey: subkey,
	}

	var reply BlobArg

	// Synchronous call

	err = conn.Client.Call("BigBlobStore.Get", args, &reply)
	conn.InUse = false

	if err != nil {
		return nil, err
	}

	return &reply, nil
}

func PostBytes(key string, subkey string, data *([]byte)) (*BlobArg, error) {

	con, err := GetConnection(":8008")
	if err != nil {
		return nil, err
	}
	// Synchronous call
	args := &BlobArg{
		Key:     key,
		SubKey:  subkey,
		Payload: *data,
	}

	var reply BlobArg
	err = con.Client.Call("BigBlobStore.Post", args, &reply)
	con.InUse = false
	if err != nil {
		return nil, err
	}

	return &reply, nil
}
