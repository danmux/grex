package app

// pond - the connection pool manager

import (
	"errors"
	"log"
	"net/rpc"
	"sync"
)

type connection struct {
	Client *rpc.Client
	InUse  bool
	IsBad  bool
}

type pond struct {
	mu       sync.Mutex
	ConnPool map[string][]*connection
	PoolSize int
}

func (b *pond) init(poolSize int) {
	b.ConnPool = make(map[string][]*connection)
	b.PoolSize = poolSize
}

func (b *pond) getConnection(connectString string) (*connection, error) {
	// lock this so we dont have any chance of two processes attempting to use the same connection
	b.mu.Lock()

	// our pools are per url (connection string)
	conList, in := b.ConnPool[connectString]
	if !in {
		log.Println("Debug - adding new pond for > " + connectString + " to the pool map")
		conList = make([]*connection, b.PoolSize)
		b.ConnPool[connectString] = conList
	}

	// pipe is our wrapper the the tcp connection
	var pipe *connection
	var err error
	var client *rpc.Client // defined here so we can keep err in this scope

	for i, con := range conList {
		// healthy existing connection
		if con != nil && !con.IsBad {
			if !con.InUse {
				pipe = con
				pipe.InUse = true
				break
			}
		} else {
			// the connection must have been marked as bad from a previous failure - so close it
			if con != nil {
				con.Client.Close()
			}
			// now we can dial this connection, or redial if it was an existing bad connection
			client, err = rpc.DialHTTP("tcp", connectString)
			if err == nil {
				// make one of our little connection wrappers 
				pipe = &connection{
					client,
					true,
					false,
				}
				conList[i] = pipe
				break
			} else {
				err = errors.New("connection failed to: " + connectString)
				markNodeUpOrDown(connectString, false)
				break
			}
		}
	}
	b.mu.Unlock()

	if pipe == nil && err == nil {
		err = errors.New("Connection pond exhausted")
	}
	return pipe, err
}

// Get a latent connection from from the pond or create a new one and add it to the pond
func GetConnection(connectString string) (*connection, error) {
	return connectionPond.getConnection(connectString)
}
