package app

// datacache - cache of decoded data types - no need to decode each time
// any puts to the items will invalidate the cache forcing a reload

import (
	"crypto/rand"
	"encoding/hex"
	"ext/vitesse/cache"
	"io"
)

type Sesh struct {
	Key  string
	Auth bool
	Uid  string
}

func (s Sesh) Size() int {
	return 27
}

func GenerateRandomKey(strength int) *string {
	k := make([]byte, strength)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	s := hex.EncodeToString(k)
	return &s
}

// versions stored in single file per bucket - each bucket has an item hash with version numbers
var seshCache *cache.LRUCache // the lru of bucket versions

func initialiseSeshCache(size int64) {
	seshCache = cache.NewLRUCache(size)
}

func newSesh() *Sesh {
	k := GenerateRandomKey(8)
	if k == nil {
		return nil
	}

	return &Sesh{
		Key:  *k,
		Auth: false,
		Uid:  "",
	}
}

func clearSeshCache() {
	seshCache.Clear()
}

func invalidateSeshInCache(sesh *Sesh) {
	seshCache.Delete(sesh.Key)
}

// put it in the cache and persist it to disk
func PutSeshInCache(sesh *Sesh) {
	seshCache.Set(sesh.Key, *sesh)
}

// return the version map for the bucket
func GetSeshFromCache(seshKey string) (*Sesh, bool) {
	sv, in := seshCache.Get(seshKey)
	ss := sv.(Sesh)
	return &ss, in
}
