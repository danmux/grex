package app

import (
	"fmt"
	"testing"
)

func Test_AddNodeLow(t *testing.T) { //test function starts with "Test" and takes a pointer to type testing.T

	InitGrex("/what/path", "my.eg.uri", "8888", 10)
	AddNode("testurl1")

	// node index 0 must always be the current node
	v, _ := LookupNode(0)
	if v.Url != "my.eg.uri:8888" {
		t.Error("Initiluse didnt add my url at location 0. " + v.Url)
	}

	v, _ = LookupNode(1)

	if v.Url != "testurl1" {
		t.Error("AddNode didnt add at location 1.")
	} else {
		t.Log("one test passed.")
	}
}

func Test_AddNodeHigh(t *testing.T) {

	InitGrex("/what/path", "localhost-something-normally", "8888", 10)

	// add a 
	AddNode("testurl23")

	v, _ := LookupNode(1)
	if v.Url != "testurl23" {
		t.Error("AddNode didnt add at location 23.")
	} else {
		t.Log("two test passed.")
	}

	if len(farm.NodeUris) != 2 {
		t.Error("farm grew badly.")
	}

	ind, isNew, err := AddNode("testurl 30")

	if ind != 2 {
		t.Error("Bad node index.")
	}

	if !isNew {
		t.Error("Node should be new")
	}

	ind, isNew, err = AddNode("testurl 30")

	if ind != 2 {
		t.Error("Bad node index.")
	}

	if isNew {
		t.Error("Node should not be new")
	}

	v, _ = LookupNode(2)
	if v.Url != "testurl 30" {
		t.Error("AddNode didnt add at location 30.")
	} else {
		t.Log("three test passed.")
	}

	if len(farm.NodeUris) != 3 {
		t.Error("farm grew badly.")
	}

	for i := 0; i < MAX_NODES-3; i++ {
		f := fmt.Sprintf("good %d", i)
		AddNode(f)
	}

	if len(farm.NodeUris) != MAX_NODES {
		t.Error("wrong number of nodes")
	}

	_, _, err = AddNode("poo")
	if err == nil {
		t.Error("too many nodes allowed")
	}

	node, err := LookupNode(135)
	if node.Url != "good 132" {
		t.Error("wrong node at index 134")
	}

	_, err = LookupNode(13400)
	if err == nil {
		t.Error("allowed me to lookmup stupid node")
	}
}

func Test_AddNodeToFlock(t *testing.T) {
	InitGrex("/what/path", "myurl", "0000", 10) // node index 0

	AddNode("testurl 0") // node index 1
	AddNode("testurl 1")
	AddNode("testurl 2")
	AddNode("testurl 3") // will have index 4
	AddNode("testurl X")

	err := AddNodeToFlock("testurl 3", "ab", true, false)
	if err != nil {
		t.Error(err)
	}

	if farm.Farm["ab"][0].Node != 4 {
		t.Error("flock 'ab' node 0 has wrong node index")
	}

	err = AddNodeToFlock("testurl 3", "ab", true, false)
	if err == nil {
		t.Error("Same node added to flock more than once")
	}

	err = AddNodeToFlock("testurl not known", "ab", true, false)
	if err == nil {
		t.Error("node should have been rejected as it is not in our node list")
	}

	// add a new node to the ab flock
	err = AddNodeToFlock("testurl X", "ab", true, true)

	if farm.Farm["ab"][1].Node != 5 {
		t.Error("flock 1 has wrong node")
	}

	for k, fl := range farm.Farm {
		t.Log(k)
		for _, nd := range fl {
			t.Log("  >", nd.Node)
		}
	}
}
