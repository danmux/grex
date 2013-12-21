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
func InitGrex(dataRootLoc string, uri string, port string, pond int, sesh int) {

	setRootPath(dataRootLoc, port)
	farm = nodeMap{}
	farm.MyUri = uri + ":" + port

	// start with 10 but can be thousands
	farm.NodeIds = make(NodeNameMap, 10)
	farm.Farm = make(map[string]nodeList)

	// add me to the Lookups and map, im up and i might be a session server
	localNode, isNew, err := AddNode(farm.MyUri, true, sesh > 0)
	// we cant have any errors adding the local node
	if err != nil {
		log.Panic(err)
	}

	if !isNew {
		log.Panic("The local node can only be added once")
	}

	localNode.Local = true

	farm.localNode = localNode

	// initialise the app with pond connections in the pond per url
	connectionPond.init(pond)

	initialiseVersionCache(1024 * 100) // 100k for versions

	// set up the session cach in 1M steps
	if sesh > 0 {
		initialiseSeshCache(int64(sesh) * 1024 * 1024)
	}

	initialiseItemCache(300000) // 300,0000 rows - eg if its xactions - this will be about 20M

	log.Println("Local node: ", LocalNodeStatus())
}

// start serving our bleet server - and find out all other nodes from the seed list
func StartServing(bleetAddy string, restAddy string, seedUrls []string) {

	go StartBleeting(bleetAddy)

	// try the seed uris for lists of nodes
	for _, s := range seedUrls {
		err := tellNodes(s)
		if err != nil {
			log.Println(err)
		}
	}

	refreshSeshServers()

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
