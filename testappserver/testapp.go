package main

import (
	"ext/go-json-rest"
	"flag"
	"grex/app"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Xact struct {
	Description string
	Other       string
	Amount      int
	Date        int64
}

type XactList []Xact

type AccountXacts struct {
	Name  string
	Xacts XactList
}

func (ax AccountXacts) Size() int {
	return len(ax.Xacts)
}

func PostXactList(key string, itemKey string, e *XactList) (string, error) {
	app.PutItemInCache(key, itemKey, AccountXacts{"shi", *e})
	enc, m := app.GetBufferEncoder()
	enc.Encode(*e)
	data := m.Bytes()
	return app.PostBytes(key, itemKey, &data)
}

func GetXactList(key string, itemKey string, e *AccountXacts) error {

	val, in := app.GetItemFromCache(key, itemKey)
	if in {
		log.Println("Yay cache hit")
		xacts := val.(AccountXacts)
		e.Xacts = xacts.Xacts
		return nil
	}

	dec, err := app.GetLoadedDecoder(key, itemKey)
	if err != nil {
		return err
	}

	err = dec.Decode(&(e.Xacts))

	if err == nil {
		log.Println("Putting it in cache")
		app.PutItemInCache(key, itemKey, *e)
	}
	return err
}

func makeSomeXacts() *XactList {
	x1 := Xact{
		Description: "My very first description",
		Other:       "My other description",
		Amount:      1245,
		Date:        1232329394,
	}

	x2 := Xact{
		Description: "My very second description",
		Other:       "My other scond ",
		Amount:      12459,
		Date:        45343254326,
	}

	thing := make(XactList, 2000)

	for i := 0; i < 2000; i += 2 {
		x1.Description = "My Nicely seperated " + strconv.Itoa(i)
		thing[i] = x1
		thing[i+1] = x2
	}

	return &thing
}

func saveSomeXacts(thing *XactList) {

	PostXactList("danmull", "poopoo", thing)
	PostXactList("ganmull", "poopoo", thing)
	PostXactList("ianmull", "poopoo", thing)
	PostXactList("kanmull", "poopoo", thing)
	PostXactList("lanmull", "poopoo", thing)
	PostXactList("nanmull", "poopoo", thing)
	PostXactList("panmull", "poopoo", thing)
	PostXactList("ranmull", "poopoo", thing)
	PostXactList("uanmull", "poopoo", thing)
	PostXactList("wanmull", "poopoo", thing)
}

func getXacts(w *rest.ResponseWriter, req *rest.Request) {
	key := req.PathParam("key")
	itemKey := req.PathParam("itemKey")

	xacts := AccountXacts{}
	err := GetXactList(key, itemKey, &xacts)
	if err != nil {
		rest.Error(w, err.Error(), 500)
	} else {
		w.WriteJson(xacts.Xacts)
	}
}

func postXacts(w *rest.ResponseWriter, req *rest.Request) {
	log.Println("got a post")
	key := req.PathParam("key")
	itemKey := req.PathParam("itemKey")

	xacts := make(XactList, 0)
	err := req.DecodeJsonPayload(&xacts)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stat, err := PostXactList(key, itemKey, &xacts)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		if stat != "good" {
			log.Println("Warning - Not all the nodes recieved the data")
		}
		w.WriteJson(stat)
	}
}

func StartAppServer(listen string) {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/xacts/:key/:itemKey", getXacts},
		rest.Route{"POST", "/xacts/:key/:itemKey", postXacts},
	)
	http.ListenAndServe(listen, &handler)
	log.Println("App server started: ", listen)
}

func main() {

	defaultDataRoot := "~/grex/data"
	dataRoot := ""

	flgServerPort := flag.String("serverport", "8007", "port number to start on")
	flgRestPort := flag.String("restport", "8008", "port number to start on")
	flgPort := flag.String("port", "8009", "port number to start bleeter on")
	flgDnsName := flag.String("name", "localhost", "The protocol and dns address of this server - it must be unique in the cluster")
	flgDnsSeed := flag.String("seed", "localhost:8019", "The protocol and dns address of this server - it must be unique in the cluster")
	flgDataRoot := flag.String("data", defaultDataRoot, "Root folder location for this node")
	flgPondSize := flag.Int("pond", 20, "Pond connections per node")

	flag.Parse()

	if defaultDataRoot != *flgDataRoot {
		dataRoot = *flgDataRoot
	}

	// initialise the farm for this node with the nodes uri
	app.InitGrex(dataRoot, *flgDnsName, *flgPort, *flgPondSize)

	// confirm this servers uri
	log.Println("My UIR: ", app.MyUri())

	// load the farm with all possible flocks using the first two chars of the key method
	for _, c1 := range app.KEY_CHARS {
		for _, c2 := range app.KEY_CHARS {
			herd := true
			if *flgPort == "8029" {
				if c1 == 'd' {
					herd = false
				}
			}
			app.AddNodeToFlock(app.MyUri(), string(c1)+string(c2), herd, false)
		}
	}

	// 42 loops in 10 seconds = 4.2 per second
	// 160MB per second

	// if *flgPort == "8029" {

	// 	ticker := time.NewTicker(2000 * time.Millisecond)
	// 	// thingy := makeSomeXacts()
	// 	loops := 0
	// 	go func() {

	// 		select {
	// 		case <-ticker.C:
	// 			loops = 0
	// 			for loops < 50 {
	// 				log.Println("timer")

	// 				// saveSomeXacts(thingy)

	// 				xacts := AccountXacts{}
	// 				GetXactList("kanmull", "poopoo", &xacts)

	// 				println(xacts.Size())

	// 				loops++

	// 				println(loops)

	// 			}
	// 		}
	// 	}()

	// }

	seeds := []string{*flgDnsSeed}

	// serve the bleets and rest web server, and contact all seeds to get other nodes status
	go app.StartServing(":"+*flgPort, ":"+*flgRestPort, seeds)

	StartAppServer(":" + *flgServerPort)

}
