//bleet - interprocess communication 
package app

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

// all the bleet listening functions
type Bleets struct{}

func (t *Bleets) GetBlob(request *BlobArg, response *BlobArg) error {
	return GetData(request, response)
}

func (t *Bleets) PostBlob(request *BlobArg, response *BlobArg) error {
	return PostData(request, response)
}

type SingleNodeFarm struct {
	Flocks map[string]FlockStatus
}

// Rpc call to share a nodes individual farm information
// request is the url of the requesting node
func (t *Bleets) GetFarm(request string, response *SingleNodeFarm) error {
	response.Flocks = make(map[string]FlockStatus)
	for k, fl := range farm.Farm {
		f := fl[0]
		response.Flocks[k] = f
	}

	// ad the requesting node - and if it is new then request thier farm
	_, in := farm.NodeIds[request]
	if !in {
		tellFarm(request)
	}

	return nil
}

type NodeList struct {
	Nodes nodeLookup
}

// Rpc call to share a nodes individual known nodes
// request is the url of the requesting node
func (t *Bleets) GetNodes(request string, response *NodeList) error {
	response.Nodes = farm.NodeUris

	// ad the requesting node - and if it is new then request thier farm
	_, isNew, err := AddNode(request)
	if err != nil {
		return err
	} else if isNew {
		println("NOt GOT THIS                       --- " + request)
		err := tellNodes(request)
		if err != nil {
			return err
		}
		return tellFarm(request)
	}

	return nil
}

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

// al the bleeters
func tellFarm(url string) error {
	conn, err := GetConnection(url)
	if err != nil {
		return err
	}

	caller := MyUri()

	var reply SingleNodeFarm

	err = conn.Client.Call("Bleets.GetFarm", caller, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return err
	}
	// add this to our farm
	log.Println("got a good farm response from : " + url)

	addExternalFarm(url, &reply)

	return nil
}

func tellNodes(url string) error {
	conn, err := GetConnection(url)
	if err != nil {
		return err
	}

	caller := MyUri()

	var reply NodeList

	err = conn.Client.Call("Bleets.GetNodes", caller, &reply)
	conn.InUse = false

	if err != nil {
		conn.IsBad = true
		return err
	}
	// add this to our farm
	log.Println("got a good GetNodes response from : " + url)

	// add these external nodes - if the node is new a tellFarm will be triggered on the node as well 
	addExternalNodes(&reply)

	return nil
}
