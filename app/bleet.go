package app

//bleet - interprocess communication

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Bleets struct{}

// all the bleet listening functions
//----------------------------------

// remote has requested some local data
func (t *Bleets) GetBlob(request *BlobArg, response *BlobArg) error {
	response.Key = request.Key
	response.SubKey = request.SubKey
	err := getData(response)
	return err
}

// remote wants us to store some data locally
func (t *Bleets) PutBlob(request *BlobArg, response *BlobArg) error {
	err := putDataIncVersion(request)
	response.Key = request.Key
	response.SubKey = request.SubKey
	response.Message = request.Message
	return err
}

// item key - is this only used in bleeting item versions?
type Key struct {
	BucketKey string
	ItemKey   string
}

// remote is asking us for our current version of the key
func (t *Bleets) GetVersion(item *Key, response *uint64) error {

	bv := getBucketVersion(item.BucketKey)
	*response = bv.getVersion(item.ItemKey)
	return nil
}

// remote is asking us if we have the specified session
func (t *Bleets) GetSession(key *string, response *Sesh) error {

	// first lets make sure we had our cache initialised
	if seshCache != nil {
		sesh, in, err := GetSeshFromCache(*key)
		if err != nil {
			return err
		}
		if in {
			response.Auth = sesh.Auth
			response.Uid = sesh.Uid
			response.Key = sesh.Key
			return nil
		}
	}
	// send a sorry looking session :(
	response.Auth = false
	response.Uid = ""
	response.Key = ""
	return nil
}

func (t *Bleets) PutSession(sesh *Sesh, response *int) error {

	log.Println("saving session " + sesh.Key)
	// first lets make sure we had our cache initialised
	if seshCache != nil {
		err := persistInCache(sesh)
		if err != nil {
			return err
		}
		*response = 1
		return nil
	}
	*response = 0
	return nil
}

type SingleNodeFarm struct {
	IsSessionServer bool
	Flocks          map[string]FlockStatus
}

// Rpc call to share a nodes individual farm information
// request is the url of the requesting node
func (t *Bleets) GetFarm(request string, response *SingleNodeFarm) error {
	log.Println("Debug - Ive been asked for my farm", request)
	response.IsSessionServer = seshCache != nil
	response.Flocks = make(map[string]FlockStatus)
	for k, fl := range farm.Farm {
		f := fl[0]
		response.Flocks[k] = f
	}

	// ad the requesting node - and if it is new then request thier farm
	_, in := farm.NodeIds[request]
	if !in {
		go func() {
			log.Println("Debug - a new node just asked for my farm, so Im requesting their nodes", request)
			err := tellNodes(request)
			if err != nil {
				return
			}
			log.Println("Debug - and I'm requesting their farm", request)
			tellFarm(request)
		}()
	}

	return nil
}

// wrapper struct for nodeLookups
type NodeList struct {
	Nodes NodeNameMap
}

// Rpc call to share a nodes individual known nodes
// request is the url of the requesting node
func (t *Bleets) GetNodes(request *NodeStatus, response *NodeList) error {
	response.Nodes = farm.NodeIds

	// ad the requesting node - and if it is new then request thier farm
	_, isNew, err := AddNode(request.Url, request.Up, request.SeshServer, request.ClientOnly)
	if err != nil {
		return err
	} else if isNew {
		go func() {
			log.Println("Debug - a new node just asked for the nodes I know about, so Im requesting their nodes", request.Url)
			err := tellNodes(request.Url)
			if err != nil {
				return
			}
			log.Println("Debug - and I'm requesting their farm", request.Url)
			tellFarm(request.Url)
		}()
	} else {
		// we have it so lets make sure we know its up
		markNodeUpOrDown(request.Url, true)
	}

	return nil
}

// all the client bleets....
//--------------------------

// im ready to listen...
func StartBleeting(bleetListen string) {
	hearBleet := new(Bleets)
	rpc.Register(hearBleet)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", bleetListen)
	if e != nil {
		log.Fatal("bleet listen error:", e)
	}
	http.Serve(l, nil)
}

// oi you tell me all the farms you is mainly caring about
func tellFarm(url string) error {
	conn, err := GetConnection(url)
	if err != nil {
		return err
	}

	caller := MyUri()

	var reply SingleNodeFarm

	log.Println("Debug - Asking a node for their farm:", url)
	err = conn.Client.Call("Bleets.GetFarm", caller, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return err
	}

	log.Println("Debug - Got a good farm response from : " + url)

	// add this to our farm
	addExternalFarm(url, &reply)

	return nil
}

// oi you at the end of this url tell me all the nodes you are ware of
func tellNodes(url string) error {
	log.Println("Debug - tellNodes - calls GetNodes from remote:" + url)
	conn, err := GetConnection(url)
	if err != nil {
		return err
	}

	// i must be the first node
	local := LocalNodeStatus()
	if local == nil {
		return errors.New("We have not initialised our own node before contacting peers")
	}

	var reply NodeList

	err = conn.Client.Call("Bleets.GetNodes", local, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return err
	}

	log.Println("Debug - got a good GetNodes response from : " + url)
	// add these external nodes - if the node is new a tellFarm will be triggered on the node as well
	addExternalNodes(&reply)

	return nil
}

// ask a remote node for the version for this item
func tellMeYourVersion(url string, bucketKey string, itemKey string) (uint64, error) {
	conn, err := GetConnection(url)
	if err != nil {
		return 0, err
	}

	key := &Key{
		BucketKey: bucketKey,
		ItemKey:   itemKey,
	}

	var reply uint64

	err = conn.Client.Call("Bleets.GetVersion", key, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return 0, err
	}
	return reply, nil
}

// ask a remote node for any session for this key
func tellMeYourSession(url string, seshKey string) (*Sesh, error) {
	conn, err := GetConnection(url)
	if err != nil {
		return nil, err
	}

	var reply Sesh

	err = conn.Client.Call("Bleets.GetSession", &seshKey, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return nil, err
	}
	return &reply, nil
}

// tell a remote node for the version for this item
func hereIsMySession(url string, sesh *Sesh) error {
	conn, err := GetConnection(url)
	if err != nil {
		return err
	}

	var reply int

	err = conn.Client.Call("Bleets.PutSession", sesh, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return err
	}
	return nil
}
