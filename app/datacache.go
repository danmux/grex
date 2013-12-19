package app

// datacache - cache of decoded data types - no need to decode each time
// any puts to the items will invalidate the cache forcing a reload

import (
	"ext/vitesse/cache"
)

// versions stored in single file per bucket - each bucket has an item hash with version numbers
var itemCache *cache.LRUCache // the lru of bucket versions

func initialiseItemCache(size int64) {
	itemCache = cache.NewLRUCache(size)
}

func clearItemCache() {
	itemCache.Clear()
}

func invalidateItemInCache(bucketKey string, itemKey string) {
	itemCache.Delete(bucketKey + "-" + itemKey)
}

// put it in the cache and persist it to disk
func PutItemInCache(bucketKey string, itemKey string, value cache.Value) {
	itemCache.Set(bucketKey+"-"+itemKey, value)
}

// return the version map for the bucket
func GetItemFromCache(bucketKey string, itemKey string) (cache.Value, bool) {
	return itemCache.Get(bucketKey + "-" + itemKey)
}
