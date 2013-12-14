package app

// version - maintains the version cache is the node management data and functions

import (
	// "errors"
	"bytes"
	"encoding/gob"
	"ext/vitesse/cache"
	"log"
)

// allows up to 18446744073709551615 versions - one write per millisecond - 585 million years
// unit32 would only allow 49 days (before needing some wrapping logic)
// map of item keys
type itemVersionMap map[string]uint64

// interface to allow it into the vitesse cache - so it can guesse how much ram is being used
func (i itemVersionMap) Size() int {
	return len(i) * 64
}

type bucketVersion struct {
	bucketKey    string
	itemVersions itemVersionMap
}

func (b *bucketVersion) incVersion(itemKey string) {
	item, in := b.itemVersions[itemKey]
	if !in {
		item = 0
	}
	b.itemVersions[itemKey] = item + 1
}

func (b *bucketVersion) getVersion(itemKey string) uint64 {
	item, in := b.itemVersions[itemKey]
	if in {
		return item
	}
	return 0
}

func (b *bucketVersion) setVersion(itemKey string, ver uint64) {
	b.itemVersions[itemKey] = ver
}

// versions stored in single file per bucket - each bucket has an item hash with version numbers
var bucketItemVCache *cache.LRUCache // the lru of bucket versions

func initialiseVersionCache(size uint64) {
	bucketItemVCache = cache.NewLRUCache(size) // 1MB cach enough for 15 pages? - TODO get from config
}

func clearVersionCache() {
	bucketItemVCache.Clear()
}

func resetBucketVersion(bucketKey string) error {
	blob := BlobArg{
		Key:     bucketKey,
		SubKey:  ".versions",
		Message: "",
	}

	return deleteData(&blob)
}

// write it to disk
func persistBucketVersion(bv *bucketVersion) error {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	enc.Encode(bv.itemVersions) // just the map

	blob := BlobArg{
		Key:     bv.bucketKey,
		SubKey:  ".versions",
		Message: "",
		Payload: m.Bytes(),
	}

	return putData(&blob)
}

func dePersistBucketVersion(bv *bucketVersion) error {

	blob := BlobArg{
		Key:     bv.bucketKey,
		SubKey:  ".versions",
		Message: "",
	}
	err := getData(&blob)
	// no file 
	if err != nil {
		bv.itemVersions = make(map[string]uint64)
		return nil
	}

	p := bytes.NewBuffer(blob.Payload)
	dec := gob.NewDecoder(p)
	err = dec.Decode(&(bv.itemVersions))
	if err != nil {
		return err
	}
	return nil
}

// put it in the cache and persist it to disk
func putBucketVersion(bv *bucketVersion) chan error {
	bucketItemVCache.Set(bv.bucketKey, bv.itemVersions)

	ch := make(chan error)
	go func(ibv *bucketVersion, ich chan error) {
		err := persistBucketVersion(ibv)
		ich <- err
	}(bv, ch)
	return ch
}

// return the version map for the bucket
func getBucketVersion(bucketKey string) *bucketVersion {

	bv := bucketVersion{
		bucketKey: bucketKey,
	}

	val, in := bucketItemVCache.Get(bucketKey)

	if in {
		bv.itemVersions = val.(itemVersionMap)
		log.Println("Debug - Loading version meta from cache - " + bucketKey)
	} else {
		// load it from disk
		log.Println("Debug - Loading version meta from disk - " + bucketKey)
		err := dePersistBucketVersion(&bv)
		// and put it in the cache
		bucketItemVCache.Set(bv.bucketKey, bv.itemVersions)

		if err != nil {
			log.Println("Error - Corrupt version vile " + bucketKey + " so resetting version file")
			log.Println(err)
			bv.itemVersions = make(map[string]uint64)
		}
	}

	return &bv
}
