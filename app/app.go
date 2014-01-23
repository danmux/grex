// Grex - An application server cluster - sharded and redundant - with application objec cache coordination and bucket based file system backed storage
package app

import (
	"log"
)

// the tcp connection pond
var connectionPond pond

// nodes and flock management
var farm nodeMap

// initGrex from command line and conf file
func LoadGrex() {
	log.Println("Debug - Grex is flocking itself")
	err := loadConf()

	if err != nil {
		log.Println("Error - " + err.Error())
	}

	InitGrex(config.StoreRoot, config.InternalName, config.Ports.Bleeter, config.PondSize, config.SeshCacheSize)
}

// start serving useing comand line and config file
func ServeGrex() {
	StartServing(farm.MyUri, config.ExternalName+":"+config.Ports.IntRest, config.Seeds)
}

// set up the farm without a config file or command line parsing
func InitGrex(storeRootLoc string, uri string, port string, pond int, sesh int64) {

	storeRoot = storeRootLoc

	forceFolders()

	farm = nodeMap{}
	farm.MyUri = uri + ":" + port

	// start with 10 but can be thousands
	farm.NodeIds = make(NodeNameMap, 10)
	farm.Farm = make(map[string]nodeList)

	// add me to the Lookups and map, im up and i might be a session server
	localNode, isNew, err := AddNode(farm.MyUri, true, sesh > 0, config.ClientOnly)
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

	log.Println("Debug - Local node: ", LocalNodeStatus())

	log.Println("Debug - Client only: ", config.ClientOnly)

	// in future these will come from the config files
	SetupDefaultFlocks(config.ClientOnly)
}

// start serving our bleet server - and find out all other nodes from the seed list
func StartServing(bleetAddy string, restAddy string, seedUrls []string) {

	go StartBleeting(bleetAddy)

	// try the seed uris for lists of nodes
	for _, s := range seedUrls {
		log.Println("Debug - querying seeds for their nodes", s)
		err := tellNodes(s)
		if err != nil {
			log.Println("Error - " + err.Error())
		}
	}

	refreshSeshServers()

	StartRestServer(restAddy)
}
