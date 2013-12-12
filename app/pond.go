// pond - the connection pool manager
package app

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
	b.mu.Lock()
	conList, in := b.ConnPool[connectString]
	if !in {
		log.Println("adding new pool for > " + connectString + " to the pool map")
		conList = make([]*connection, b.PoolSize)
		b.ConnPool[connectString] = conList
	} else {
		log.Println("got existing pool for > " + connectString)
	}

	var pipe *connection
	var err error
	var client *rpc.Client

	for i, con := range conList {
		if con != nil && !con.IsBad {
			if !con.InUse {
				log.Printf("found free connection %d\n", i)
				pipe = con
				pipe.InUse = true
				break
			}
		} else {
			if con.IsBad {
				con.Client.Close()
			}

			log.Printf("creating a new connection %d\n", i)
			client, err = rpc.DialHTTP("tcp", connectString)
			if err == nil {
				log.Println("good new connection")
				pipe = &connection{
					client,
					true,
					false,
				}
				conList[i] = pipe
				break
			} else {
				err = errors.New("connection failed to: " + connectString)
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

var connectionPond pond

// Get a latent connection from from the pond or create a new one and add it to the pond
func GetConnection(connectString string) (*connection, error) {
	return connectionPond.getConnection(connectString)
}

// Set up the connection pond with the given size per url
func InitPond(size int) {
	connectionPond.init(size)
}
