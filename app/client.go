package app

// used to encode client types to anf from blobs to send across the farm

import (
	"bytes"
	"encoding/gob"
	"ext/vitesse/cache"
	"log"
)

type BlobObj interface{}

// recover from the cluster into obj the object at key, itemkey 
func GetObject(key string, itemKey string, obj BlobObj) error {

	dec, err := GetLoadedDecoder(key, itemKey)
	if err != nil {
		return err
	}

	err = dec.Decode(obj)

	return err
}

// store this object in the cluster
func PutObject(key string, itemKey string, obj BlobObj) (string, error) {

	enc, m := GetBufferEncoder()
	enc.Encode(obj)
	data := m.Bytes()

	return PostBytes(key, itemKey, &data)
}

// cache this object tore this object in the cluster
func PutCachedObject(key string, itemKey string, obj cache.Value) (string, error) {
	msg, err := PutObject(key, itemKey, obj)
	PutItemInCache(key, itemKey, obj)
	return msg, err
}

func GetCachedObject(key string, itemKey string, val cache.Value) (cache.Value, error) {

	obj, in := GetItemFromCache(key, itemKey)
	if in {
		log.Println("Debug - Got data cache hit", key, itemKey)
		return obj, nil
	}

	dec, err := GetLoadedDecoder(key, itemKey)
	if err != nil {
		return val, err
	}

	err = dec.Decode(val)

	if err == nil {

		log.Println("Debug - storing in cache", key, itemKey, val)

		PutItemInCache(key, itemKey, val)
		return val, nil
	}

	return val, err
}

// Return a GOB decoder primed wiht the bytes for a given key and subkey 
func GetLoadedDecoder(key string, subkey string) (*gob.Decoder, error) {
	reply, err := GetBytes(key, subkey)

	if reply.Message == "missing" {
		return nil, err
	}

	p := bytes.NewBuffer(reply.Payload)
	//bytes.Buffer satisfies the interface for io.Writer and can be used
	//in gob.NewDecoder() 
	return gob.NewDecoder(p), nil
}

// Return an encoder and buffer ready to encode anything
func GetBufferEncoder() (*gob.Encoder, *bytes.Buffer) {
	// some optimisations for bulk data would be to prealocate the buffer and cache the encoder
	m := new(bytes.Buffer)
	//the *bytes.Buffer satisfies the io.Writer interface and can
	//be used in gob.NewEncoder() 
	enc := gob.NewEncoder(m)

	return enc, m
}
