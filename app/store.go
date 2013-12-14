package app

import (
	"errors"
	"log"
)

type BlobArg struct {
	Key     string
	SubKey  string
	Message string
	Payload []byte
}

type UrlAndVer struct {
	url string
	ver uint64
}

// for all herders contact them in parallel to get the versions return the  
func getRemoteVersions(herders []string, key string, itemKey string) (uint64, *([]string)) {

	N := len(herders)

	// if we only have the local node then return any previous error
	if N == 0 {
		return 0, nil
	}

	sem := make(chan UrlAndVer)

	// and send it to all herding nodes in parallel
	for _, nodeUrl := range herders {

		go func(url string, key string, itemKey string) {
			ver, err := tellVersion(url, key, itemKey)
			if err != nil {
				log.Println("Error - failed to get version from node: " + url)
				ver = 0
			}

			sem <- UrlAndVer{
				url: url,
				ver: ver,
			}
		}(nodeUrl, key, itemKey)
	}

	maxRemoteVersion := uint64(0)
	var uptoDateUrls []string
	// wait for goroutines to finish
	for i := 0; i < N; i++ {
		node := <-sem
		if node.ver > maxRemoteVersion {
			uptoDateUrls = make([]string, 0)
			uptoDateUrls = append(uptoDateUrls, node.url)
			maxRemoteVersion = node.ver
		} else if node.ver == maxRemoteVersion {
			uptoDateUrls = append(uptoDateUrls, node.url)
		} else {
			log.Println("Warning - some nodes are out of date")
		}
	}

	return maxRemoteVersion, &uptoDateUrls
}

// Any node can recieve a request to read bytes
// If this node does not have this item and should have - it will gather it from the first node that does have the item (read repair)
// First we contact all replicating nodes for this item to ask for their version, it finds the maximum version, persists that locally (repair)
// recording the version number - it should also send this max version to the other nodes

func GetBytes(key string, itemKey string) (*BlobArg, error) {

	args := &BlobArg{
		Key:    key,
		SubKey: itemKey,
	}
	reply := args

	var err error
	errorCount := 0
	// get whether im herding it and the list of other herders - allow reads from a partial flock
	imHerding, othersHerding := getHerdersForBucket(key, true)
	if !imHerding && len(othersHerding) == 0 {
		reply.Message = "critical"
		return reply, errors.New("No nodes registered for this bucket: " + key)
	}

	// the consistancy flag
	imInconsistant := false

	// get max remote versions - perhaps add an opting to not do this - for fast mode?
	maxRemoteVersion, uptoDateUrls := getRemoteVersions(othersHerding, key, itemKey)

	var bv *bucketVersion

	// do we care about this bucket - is it in a flock we are herding - then grab it locally
	// local disk access is still quicker than getting it from a remote nodes cache?
	if imHerding {
		log.Println("im hearding :" + key)

		// first check local version if it is less than the max remote versions 
		// then get dont get this now - and set the should repair flag - so we persist the latest version locally
		bv := getBucketVersion(key)
		myVer := bv.getVersion(itemKey)

		if myVer < maxRemoteVersion {
			log.Printf("local inconsistant with version %d remotes as %d\n", myVer, maxRemoteVersion)
			imInconsistant = true
		} else {
			log.Println("Debug - local version is latest")
			err = getData(args)
			reply = args
			if err != nil {
				errorCount++
				log.Println("Warning - Reading from local disk failed - attempting to recover from replicas")
			} else {
				// got a good local copy so return with that
				return reply, nil
			}
		}
	}

	// synchronously try each node that have the highest version numbers in turn - until one succeeds
	// TODO - prioritise quiet nodes
	for _, nodeUrl := range *uptoDateUrls {
		conn, err := GetConnection(nodeUrl)
		if err != nil {
			errorCount++
			log.Println("Error - failed to get connection for: " + nodeUrl)
		} else {

			err = conn.Client.Call("Bleets.GetBlob", args, reply)
			conn.InUse = false
			if err != nil {
				conn.IsBad = true
				errorCount++
				log.Println("Error - failed to get blob from replicating node: " + nodeUrl)
			} else {
				// got a good one from remote source, so that will do

				// so if our node was not consistant then save this file and version
				if imInconsistant {
					log.Printf("Got inconsistant local version - so updating to remote version %d", maxRemoteVersion)
					bv.setVersion(itemKey, maxRemoteVersion)
					putBucketVersion(bv)
					putData(reply)
				}
				break
			}
		}
	}

	if errorCount > 0 {
		reply.Message = "warning"

		if errorCount == len(othersHerding) {
			reply.Message = "critical"
			return reply, errors.New("Could not read this data from any node")
		}
	}

	return reply, nil
}

// get current version from version cache, up the version persist version ans persist the data,  
func PostBytes(key string, itemKey string, data *([]byte)) (string, error) {

	// get whether im herding it and the list of other herders
	imHerding, othersHerding := getHerdersForBucket(key, true)

	if !imHerding && len(othersHerding) == 0 {
		return "critical", errors.New("No nodes registered for this bucket " + key)
	}

	// how many nodes did not accept the blob
	errorCount := 0

	// our general binary payload with routing info - Key and SubKey
	args := &BlobArg{
		Key:     key,
		SubKey:  itemKey,
		Payload: *data,
	}

	var err error
	// do we care about this bucket - is it in a flock we are herding - then save it locally
	if imHerding {
		err = putDataIncVersion(args)
		if err != nil {
			errorCount++
			log.Println("Warning - Writing to local disk failed - attempting to send to replicas")
		}
	}

	// lets spread the love....
	N := len(othersHerding)

	// if we only have the local node then return any previous error
	if N == 0 {
		return "good", err
	}

	// send to others asynchronously
	sem := make(chan int)
	// store the results
	res := make([]BlobArg, N)
	// sore our connections so we can return them to the pond  
	cons := make([]*connection, N)

	// and send it to all herding nodes
	for ix, nodeUrl := range othersHerding {
		cons[ix], err = GetConnection(nodeUrl)
		if err != nil {
			errorCount++
			log.Println("Error - failed to get connection for: " + nodeUrl)
		} else {
			go func(i int) {
				err := cons[i].Client.Call("Bleets.PutBlob", args, &res[i])
				// return then to the pond
				cons[i].InUse = false
				if err != nil {
					cons[i].IsBad = true
					errorCount++
					log.Println("Error - failed to send blob to replicating node: " + nodeUrl)
				}
				sem <- i
			}(ix)
		}
	}

	// wait for goroutines to finish
	for i := 0; i < N; i++ {
		<-sem
	}

	// and work out what happened...
	resp := "good"

	if errorCount > 0 {
		resp = "warning"

		if errorCount == len(othersHerding)+1 {
			return "critical", errors.New("NO nodes persisted the data")
		}
		if errorCount == len(othersHerding) {
			return "error", errors.New("Only the local node persisted the blob")
		}
	}

	return resp, nil
}
