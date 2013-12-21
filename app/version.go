package app

// version - maintains the version cache all buckets have a .versions file that keeps track of versoions of 
//           all keys in the bucket - all writes update the file on disk and ripples it across any herder node
//           updating thier versions.
// 			 versions works hand in hand withe the store...
//           A read of the version file will only attempt to get the version from cache or the local disk - not the cluster.  
//           the data reads themselves (from the store) will have contacted all herders to check their versions.
//           and updated the local copy to the latest version if needed.

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

// interface to allow it into the vitesse cache - so it can guess how much ram is being used
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

func initialiseVersionCache(size int64) {
	bucketItemVCache = cache.NewLRUCache(size)
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

// return the version map for the bucket - from our version cache first and foremost
// then from our local persistant serialisation store (or disk as it used to be known)
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

		if err != nil || bv.itemVersions == nil {
			log.Println("Error - Corrupt version vile " + bucketKey + " so resetting version file")
			if err != nil {
				log.Println(err)
			}
			bv.itemVersions = make(map[string]uint64)
		}
		// and put it in the cache
		bucketItemVCache.Set(bv.bucketKey, bv.itemVersions)
	}

	return &bv
}
