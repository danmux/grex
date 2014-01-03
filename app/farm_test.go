package app

import (
	"fmt"
	"testing"
)

func Test_AddNodeLow(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	InitGrex("../testdata", "my.eg.uri", "8888", 10, 0)
	AddNode("testurl1", true, true)

	// node index 0 must always be the current node
	v := LocalNodeStatus()
	if v.Url != "my.eg.uri:8888" {
		t.Error("Initiluse didnt add my url at location 0. " + v.Url)
	}
}

func Test_AddNodeHigh(t *testing.T) {

	InitGrex("../testdata", "localhost-something-normally", "8888", 10, 1)

	// add a 
	AddNode("testurl23", true, true)

	if len(farm.NodeIds) != 2 {
		t.Error("farm grew badly.")
	}

	node, isNew, err := AddNode("testurl 30", true, true)

	if node == nil {
		t.Error("Bad node.")
	}

	if !isNew {
		t.Error("Node should be new")
	}

	node, isNew, err = AddNode("testurl 30", true, true)

	if node == nil {
		t.Error("Bad node index.")
	}

	if isNew {
		t.Error("Node should not be new")
	}

	if len(farm.NodeIds) != 3 {
		t.Error("farm grew badly.")
	}

	for i := 0; i < MAX_NODES-3; i++ {
		f := fmt.Sprintf("good %d", i)
		AddNode(f, true, true)
	}

	if len(farm.NodeIds) != MAX_NODES {
		t.Error("wrong number of nodes")
	}

	_, _, err = AddNode("poo", true, true)
	if err == nil {
		t.Error("too many nodes allowed")
	}
}

func Test_AddNodeToFlock(t *testing.T) {
	InitGrex("../testdata", "myurl", "0000", 10, 2) // node index 0

	AddNode("testurl 0", true, true) // node index 1
	AddNode("testurl 1", true, true)
	AddNode("testurl 2", true, true)
	AddNode("testurl 3", true, true) // will have index 4
	AddNode("testurl X", true, true)

	err := AddNodeToFlock("testurl 3", "ab", true)
	if err != nil {
		t.Error(err)
	}

	err = AddNodeToFlock("testurl 3", "ab", true)
	if err == nil {
		t.Error("Same node added to flock more than once")
	}

	err = AddNodeToFlock("testurl not known", "ab", true)
	if err == nil {
		t.Error("node should have been rejected as it is not in our node list")
	}

	// add a new node to the ab flock
	err = AddNodeToFlock("testurl X", "ab", true)

	// for k, fl := range farm.Farm {
	// 	t.Log(k)
	// 	for _, nd := range fl {
	// 		t.Log("  >", nd.Node)
	// 	}
	// }
}
