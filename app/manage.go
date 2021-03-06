package app

// rest - the restfull grex management api
// an users app should start their own web server if (if it is an app server app) to avoid path conflicts

import (
	"encoding/json"
	"ext/go-json-rest"
	"log"
	"net/http"
)

func getFarm(w *rest.ResponseWriter, req *rest.Request) {
	farm := Farm()
	w.WriteJson(farm)
}

func getVCache(w *rest.ResponseWriter, req *rest.Request) {
	w.Write([]byte(bucketItemVCache.StatsJSON()))
}

func getICache(w *rest.ResponseWriter, req *rest.Request) {
	w.Write([]byte(itemCache.StatsJSON()))
}

func getMeta(w *rest.ResponseWriter, req *rest.Request) {
	key := req.PathParam("blobkey")
	log.Println(key)
	bv := getBucketVersion(key)
	log.Println(bv)

	b, err := json.MarshalIndent(bv.itemVersions, "", "  ")

	if err != nil {
		log.Println(err)
		return
	}
	w.Write(b)
}

func StartRestServer(listen string) {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/.farm", getFarm},
		rest.Route{"GET", "/.vcache", getVCache},
		rest.Route{"GET", "/.icache", getICache},
		rest.Route{"GET", "/.meta/:blobkey", getMeta},
	)
	http.ListenAndServe(listen, &handler)
	log.Println("Admin rest server started: ", listen)
}
