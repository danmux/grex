package app

// seshcache - an in memory distributed session cache

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"ext/vitesse/cache"
	"io"
	"log"
)

type Sesh struct {
	Key  string `json:"key"`
	Auth bool   `json:"auth"`
	Uid  string `json:"uid"`
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

var seshServersUrls []string

func refreshSeshServers() {
	if seshCache == nil {
		return
	}
	seshServersUrls = make([]string, 0, 10)
	for _, stat := range farm.NodeIds {
		if stat.SeshServer && !stat.Local {
			seshServersUrls = append(seshServersUrls, stat.Url)
		}
	}
}

func NewSesh() *Sesh {
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

// just put it in the local cache
func persistInCache(sesh *Sesh) error {
	if seshCache == nil {
		return errors.New("This is not a session server")
	}
	seshCache.Set(sesh.Key, *sesh)
	return nil
}

// put it in the cache and share it
func PutSeshInCache(sesh *Sesh) error {
	err := persistInCache(sesh)
	if err != nil {
		return err
	}

	N := len(seshServersUrls)

	// if we only have the local node then return any previous error
	if N == 0 {
		log.Println("Warning - No other session servers")
		return nil
	}

	// send to others asynchronously
	sem := make(chan int)

	// and send it to all herding nodes
	for _, nodeUrl := range seshServersUrls {
		if err != nil {
			log.Println("Error - PutSeshInCache failed to get connection for: " + nodeUrl)
		} else {
			go func(url string) {

				err := hereIsMySession(url, sesh)

				if err != nil {
					log.Println(err)
					log.Println("Error - failed to send session to node: " + nodeUrl)
				}
				sem <- 0
			}(nodeUrl)
		}
	}

	// wait for goroutines to finish
	for i := 0; i < N; i++ {
		<-sem
	}

	return nil
}

// return the session for the given key and whether it was found
func GetSeshFromCache(seshKey string) (*Sesh, bool, error) {
	if seshCache == nil {
		return nil, false, errors.New("This is not a session server")
	}

	sv, in := seshCache.Get(seshKey)
	if in {
		ss := sv.(Sesh)
		return &ss, true, nil
	} else {
		// see if it is in remote caches eg if this node went down and up again - read repair
		N := len(seshServersUrls)

		if N == 0 {
			return nil, false, nil
		}

		sem := make(chan *Sesh)

		// and send it to all herding nodes in parallel - but well grab the first one back
		for _, nodeUrl := range seshServersUrls {

			go func(url string, key string) {
				sesh, err := tellMeYourSession(url, key)
				if err != nil {
					log.Println("Error - failed to get session from node: " + url)
				}

				sem <- sesh
			}(nodeUrl, seshKey)
		}

		var firstSesh *Sesh
		// loop round and get first none null session
		for i := 0; i < len(seshServersUrls); i++ {
			// wait for first one back
			firstSesh = <-sem
			if firstSesh != nil {
				break
			}
		}

		if firstSesh == nil {
			log.Println("Warning - got no session container for this key")
			return nil, false, nil
		}

		if firstSesh.Key == "" {
			log.Println("Warning - got no session key for this key")
			return nil, false, nil
		}

		// stick it in our local cach
		persistInCache(firstSesh)
		return firstSesh, true, nil

	}
	return nil, false, nil
}
