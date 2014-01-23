package app

// used to encode client types to anf from blobs to send across the farm

import (
	"bytes"
	"encoding/gob"
	"ext/vitesse/cache"
	"fmt"
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

		log.Println("Debug - storing in cache", key, itemKey)

		PutItemInCache(key, itemKey, val)
		return val, nil
	}

	return val, err
}

type Index struct {
	Key   string
	Index int
}

func (a Index) Size() int {
	return 10
}

func indexedFileName(itemKey string, index int) string {
	return fmt.Sprintf("%v-%05d", itemKey, index)
}

func indexName(itemKey string) string {
	return "." + itemKey + "-index"
}

func updateIndex(key string, itemKey string, index int) (int, string, error) {
	// get the index file
	ifn := indexName(itemKey)
	ind := Index{}
	iv, _ := GetCachedObject(key, ifn, &ind)
	ix := iv.(*Index)

	ix.Key = itemKey

	if index == 0 {
		// add this to the index
		ix.Index = ix.Index + 1
	} else {
		if index > ix.Index {
			ix.Index = index
		}
	}
	// put the index file
	str, err := PutCachedObject(key, ifn, ix)
	return ix.Index, str, err
}

func PostIndexedObject(key string, itemKey string, obj cache.Value) (int, string, error) {

	index, msg, err := updateIndex(key, itemKey, 0)
	if msg == "critical" {
		return 0, msg, err
	}

	fn := indexedFileName(itemKey, index)

	// put the object
	mes, err := PutCachedObject(key, fn, obj)
	return index, mes, err
}

func PutIndexedObject(key string, itemKey string, index int, obj cache.Value) (string, error) {

	fn := indexedFileName(itemKey, index)

	// put the object
	return PutCachedObject(key, fn, obj)
}

func GetIndexedObject(key string, itemKey string, index int, val cache.Value) (cache.Value, error) {
	fn := indexedFileName(itemKey, index)
	return GetCachedObject(key, fn, val)
}

func GetIndexKeys(key string, itemKey string) ([]int, error) {
	// get the index file
	ifn := indexName(itemKey)
	ind := Index{}
	iv, _ := GetCachedObject(key, ifn, &ind)
	ix := iv.(*Index)
	fkey := make([]int, 0, 5)
	for i := 0; i < ix.Index; i++ {
		fkey = append(fkey, i+1)
	}
	return fkey, nil
}

func ResetIndexes(key string, itemKey string) error {
	blob := BlobArg{
		Key:     key,
		SubKey:  indexName(itemKey),
		Message: "",
	}

	return deleteData(&blob)
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
