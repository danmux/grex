package main

import (
	"flag"
	"grex/app"
	"log"
	"time"
)

func main() {

	flgRestPort := flag.String("restport", "8008", "port number to start on")
	flgPort := flag.String("port", "8009", "port number to start bleeter on")
	flgDnsName := flag.String("name", "localhost", "The protocol and dns address of this server - it must be unique in the cluster")
	flgDnsSeed := flag.String("seed", "localhost:8019", "The protocol and dns address of this server - it must be unique in the cluster")

	flag.Parse()

	app.InitPond(20)

	app.InitFarm(*flgDnsName + ":" + *flgPort)

	// confirm this servers uri
	log.Println("My UIR: ", app.MyUri())
	// and confirm it is node 0
	node0, _ := app.LookupNode(0)
	log.Println("Node 0: ", node0)

	// load the farm with all possible flocks using the first two chars of the key method
	for _, c1 := range app.KEY_CHARS {
		for _, c2 := range app.KEY_CHARS {
			app.AddNodeToFlock(app.MyUri(), string(c1)+string(c2), true, false)
		}
	}

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("timer")
			}
		}
	}()

	seeds := []string{*flgDnsSeed}

	app.StartServing(":"+*flgPort, ":"+*flgRestPort, seeds)

}
