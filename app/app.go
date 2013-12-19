// Grex - An application server cluster - sharded and redundant - with application objec cache coordination and bucket based file system backed storage
package app

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

// the tcp connection pond
var connectionPond pond

// nodes and flock management
var farm nodeMap

// where all data is to be kept
var dataRoot string

// set up the farm 
func InitGrex(dataRootLoc string, uri string, port string, pond int) {

	setRootPath(dataRootLoc, port)
	farm = nodeMap{}
	farm.MyUri = uri + ":" + port

	farm.NodeUris = make(nodeLookup, 0, 10)
	farm.NodeIds = make(nodeNameMap)
	farm.Farm = make(map[string]nodeList)

	// add me to the Lookups and map
	AddNode(farm.MyUri)

	// initialise the app with pond connections in the pond per url
	connectionPond.init(pond)

	initialiseVersionCache(1024 * 100) // 100k for versions

	initialiseItemCache(300000) // 300,0000 rows - eg if its xactions - this will be about 20M

	// and confirm it is node 0
	node0, _ := LookupNode(0)
	log.Println("Node 0: ", node0)
}

func StartServing(bleetAddy string, restAddy string, seedUrls []string) {

	go StartBleeting(bleetAddy)

	// try the seed uris for lists of nodes
	for _, s := range seedUrls {
		err := tellNodes(s)
		if err != nil {
			log.Println(err)
		}
	}

	StartRestServer(restAddy)
}

// set the root data folder where the data and other config will be kept
func setRootPath(root string, port string) {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	// using default setting - include port to make testing on the same host less error prone
	if root == "" {
		dataRoot = usr.HomeDir + "/grex/data/" + port
	} else {
		dataRoot = root
	}

	// tidy it up and enforce trailing space
	dataRoot = filepath.Clean(dataRoot) + "/"

	if dataRoot == "//" {
		log.Fatal("You cant use the root folder for Grex data")
	}

	// ensure the data root folder exists
	err = os.MkdirAll(dataRoot, 0750)
	if err != nil {
		log.Fatal("Error - cant make root directory")
	}
}
