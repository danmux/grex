package app

// farm - is the node management data and functions

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

// a farm has a list of these known nodes
type NodeStatus struct {
	Url        string
	Up         bool
	SeshServer bool // is it a session server
	Local      bool // is this the local node
}

// node string uri (which will almost certainly be a url as well) to nodeIndex lookup
type NodeNameMap map[string](*NodeStatus)

// for a particular flock what does the Node (givien by the node index) care
type FlockStatus struct {
	Node *NodeStatus
	// does the Node want to herd this flock
	Herder bool
	// is if herding - if this is false and Herder is true, then the Node has not got a copy of the flock data yet
	Herding bool
}

// array (rather than linked list as modes rarely added or removed)
type nodeList []FlockStatus

// The nodeMap contains the 
type nodeMap struct {
	MyUri string

	localNode *NodeStatus

	NodeIds NodeNameMap
	// a farm is a map of all our flocks and the list of nodes serving each flock
	Farm map[string]nodeList
}

// return the uri for this node
func Farm() *nodeMap {
	return &farm
}

// return the uri for this node
func MyUri() string {
	return farm.MyUri
}

// this is the flock key algorithm - default simply the first two characters of the bucket key
func getFlockKey(bucketKey string) string {
	return bucketKey[0:2]
}

// for a given bucket return whether the current node is herding it and a list of other node urls that are herding it as well
// flag to rule out those that are not actually fully herding yet
func getHerdersForBucket(bucketKey string, allowPartialHerding bool) (bool, []string) {
	flockKey := getFlockKey(bucketKey)
	iCare := false
	herders := make([]string, 0, MAX_REPLICAS)
	for _, fs := range farm.Farm[flockKey] {
		if fs.Node.Local {
			iCare = fs.Herding || (allowPartialHerding && fs.Herder)
		} else if fs.Herder {
			if allowPartialHerding || fs.Herding {
				nodeStatus := fs.Node
				if nodeStatus == nil {
					log.Println("Error - inconsistant farm")
				} else if !nodeStatus.Up {
					log.Println("Warning - wanted to send stuff to a downed node")
				} else {
					herders = append(herders, nodeStatus.Url)
				}
			}
		}
	}
	return iCare, herders
}

// Register that a Node with given uri is herding / or not a particular flock 
func AddNodeToFlock(uri string, flockKey string, herder bool) error {
	// check the Node exists in our list
	nodeStat, in := farm.NodeIds[uri]
	if !in {
		return errors.New("Tried to add an unknown node " + uri + " to flock: " + flockKey)
	}

	flock := FlockStatus{
		Node:   nodeStat,
		Herder: herder,
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
// returns the new node status, and if it is new
func AddNode(uri string, up bool, sesh bool) (*NodeStatus, bool, error) {

	// got this one already
	nodeStat, in := farm.NodeIds[uri]
	if in {
		return nodeStat, false, nil
	}

	// make sure we dont excede the maximum number of known nodes
	if len(farm.NodeIds) >= MAX_NODES {
		return nil, false, errors.New("Too many nodes for the world")
	}

	// otherwise make a new node and add it to the list
	newNode := NodeStatus{uri, up, sesh, false}
	// and our per uri node status lookup
	farm.NodeIds[uri] = &newNode
	return &newNode, true, nil
}

func addExternalFarm(url string, newfarm *SingleNodeFarm) {
	for k, f := range newfarm.Flocks {
		AddNodeToFlock(url, k, f.Herder)
	}
}

// for any nodes found in the recently acquired nodelist if they are new then get that nodes farm status as well
func addExternalNodes(newNodes *NodeList) {
	for _, f := range newNodes.Nodes {
		_, isNew, err := AddNode(f.Url, f.Up, f.SeshServer)
		if err == nil {
			// any new nodes go and get the farm status
			if isNew {
				tellFarm(f.Url)
			}
		}
	}
}

// report a node as down when we cant get a connection to it
func markNodeUpOrDown(url string, upOrDown bool) {
	nodeStatus, in := farm.NodeIds[url]
	if in {
		nodeStatus.Up = upOrDown
	}
}

// load the farm with all possible flocks using the first two chars of the key method
func SetupDefaultFlocks() {
	for _, c1 := range KEY_CHARS {
		for _, c2 := range KEY_CHARS {
			herd := true
			AddNodeToFlock(MyUri(), string(c1)+string(c2), herd)
		}
	}
}

func LocalNodeStatus() *NodeStatus {
	return farm.localNode
}
