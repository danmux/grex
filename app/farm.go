// farm - is the node management data functions
package app

import (
	"errors"
	"log"
)

// the maximum number of nodes in the cluster 
// Arbitrarily set to the same number of flocks in the two character flocking system
const MAX_NODES = 1296

// Each flock can have at most 10 replicas
const MAX_REPLICAS = 10

// The set of allowable characters in any key
const KEY_CHARS = "01234567890abcdefghijklmnopqrstuvwxyz"

type nodeIndex uint16 // 65535 nodes maximum (2 per flock)

type NodeStatus struct {
	Url string
	Up  bool
}

// node nodeIndex to string lookup 
type nodeLookup []NodeStatus

// node string uri (which will almost certainly be a url as well) to nodeIndex lookup
type nodeNameMap map[string]nodeIndex

// for a particular flock
type FlockStatus struct {
	Node   nodeIndex // so our nodemap does not contain strings - just integer references to the nodeLookup
	Herder bool
	Cached bool
}

// array (rather than linked list as modes rarely added or removed)
type nodeList []FlockStatus

// The nodeMap contains the 
type nodeMap struct {
	MyUri string
	// the nodeIndex to node Uri mapping thing
	NodeUris nodeLookup
	NodeIds  nodeNameMap
	// a farm is a map of all our flocks 
	Farm map[string]nodeList
}

var farm nodeMap

// return the uri for this node
func Farm() *nodeMap {
	return &farm
}

// return the uri for this node
func MyUri() string {
	return farm.MyUri
}

// set up the farm 
func InitFarm(uri string) {

	farm = nodeMap{}
	farm.MyUri = uri

	farm.NodeUris = make(nodeLookup, 0, 10)
	farm.NodeIds = make(nodeNameMap)
	farm.Farm = make(map[string]nodeList)

	// add me to the Lookups and map
	AddNode(uri)
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

// Register that a Node with given uri is herding / or not a particular flock 
func AddNodeToFlock(uri string, flockKey string, herder bool, cached bool) error {
	// check the Node exists in our list
	index, in := farm.NodeIds[uri]
	if !in {
		return errors.New("Tried to add an unknown node " + uri + " to flock: " + flockKey)
	}

	flock := FlockStatus{
		Node:   index,
		Herder: herder,
		Cached: cached,
	}

	nl, in := farm.Farm[flockKey]
	if !in {
		nl = make(nodeList, 0, MAX_REPLICAS)
	}

	// check if this node index is already in
	for _, f := range nl {
		if f.Node == flock.Node {
			// is this error needed really - any harm trying to add the same node?
			return errors.New("this node has allready been added to that flock")
		}
	}

	if len(nl) >= MAX_REPLICAS {
		return errors.New("Too many nodes for this flock")
	}
	nl = append(nl, flock)
	farm.Farm[flockKey] = nl
	return nil
}

// Add a node uri to the node lookups
// returns the index for the node, and if it is new
func AddNode(uri string) (nodeIndex, bool, error) {

	// got this one already
	id, in := farm.NodeIds[uri]
	if in {
		return id, false, nil
	}

	if len(farm.NodeUris) >= MAX_NODES {
		return 0, false, errors.New("Too many nodes for the world")
	}

	// otherwise ad it to the list
	farm.NodeUris = append(farm.NodeUris, NodeStatus{uri, true})
	id = nodeIndex(len(farm.NodeUris) - 1)

	// and our index lookup
	farm.NodeIds[uri] = id
	return id, true, nil
}

// Return the node uri for a given internal node index
func LookupNode(id nodeIndex) (NodeStatus, error) {
	if id >= nodeIndex(len(farm.NodeUris)) {
		return NodeStatus{}, errors.New("that node id does not exist in our node list")
	}
	return farm.NodeUris[id], nil
}

func addExternalFarm(url string, newfarm *SingleNodeFarm) {
	for k, f := range newfarm.Flocks {
		AddNodeToFlock(url, k, f.Herder, f.Cached)
	}
}

// for any nodes found in the recently acquired nodelist if they are new then get that nodes farm status as well
func addExternalNodes(newNodes *NodeList) {
	for _, f := range newNodes.Nodes {
		ind, isNew, err := AddNode(f.Url)
		if err == nil {
			farm.NodeUris[ind].Up = f.Up
			// any new nodes go and get the farm status
			if isNew {
				tellFarm(f.Url)
			}
		}
	}
}
