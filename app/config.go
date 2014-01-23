package app

// config - read in the given config file apply any command line overrides

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

type Ports struct {
	Bleeter string `json:"bleet"`
	IntRest string `json:"rest_internal"`
	ExtRest string `json:"rest_external"` // if this is a restfull app server
}

type Config struct {
	InternalName  string   `json:"internal_ip"` // the internal ip to bind to for bleeting
	ExternalName  string   `json:"external_ip"` // external for the rest interfaces
	Ports         Ports    `json:"ports"`
	Seeds         []string `json:"seeds"`           // the peer servers in this cluster
	StoreRoot     string   `json:"store_root"`      // the root file location
	PondSize      int      `json:"pond_size"`       // how big the pond is per node
	SeshCacheSize int64    `json:"sesh_cache_size"` // how big the session cache size - number of sessions
	DataCacheSize int64    `json:"data_cache_size"` // how big the data cache size is - number of items (of average size)
	ClientOnly    bool     `json:"client_only"`     // default false, but is true then it wont flock any data
}

var config Config
var confFilePath string

var defaultPorts = Ports{
	Bleeter: "8006",
	IntRest: "8007",
	ExtRest: "8008",
}

var defaultSeeds = []string{"localhost:8016"}

var defConf = Config{
	InternalName:  "localhost",
	ExternalName:  "localhost",
	Ports:         defaultPorts,
	Seeds:         defaultSeeds,
	StoreRoot:     "~/grex/store",
	PondSize:      20,
	SeshCacheSize: 1,      //1M
	DataCacheSize: 300000, //row items
}

// where all data is to be kept
var storeRoot string
var dataRoot string // the data files

// set the root data folder where the data and other config will be kept
func setRootPath(root string, port string) {

	// using default setting - include port to make testing on the same host less error prone
	if root == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		storeRoot = usr.HomeDir + "/grex/store/" + port
	} else {
		storeRoot = root
	}
}

func parseFlags() (*Config, bool) {

	clConf := Config{}
	clPorts := Ports{}

	flag.StringVar(&confFilePath, "conf", "", "Json config file name must have .json extension")

	flag.StringVar(&clConf.StoreRoot, "data", defConf.StoreRoot, "Root folder location for this node")

	flag.StringVar(&clConf.InternalName, "bindinternal", defConf.InternalName, "Bind to internal ip it must be unique in the cluster")
	flag.StringVar(&clConf.ExternalName, "bindexternal", defConf.ExternalName, "Bind to external ip it must be unique in the cluster")

	flag.StringVar(&clPorts.ExtRest, "serverport", defConf.Ports.ExtRest, "port number to start any external app server on")
	flag.StringVar(&clPorts.IntRest, "managerport", defConf.Ports.IntRest, "port number to start internal/external rest management server on")
	flag.StringVar(&clPorts.Bleeter, "bleeterport", defConf.Ports.Bleeter, "Port number to start internal rpc bleeter on")

	flgDnsSeed := flag.String("seed", defConf.Seeds[0], "The protocol and dns address of this server - it must be unique in the cluster")

	flag.IntVar(&clConf.PondSize, "pond", defConf.PondSize, "Pond connections per node")
	flag.Int64Var(&clConf.SeshCacheSize, "seshcache", defConf.SeshCacheSize, "integer value of megabytes to assign to the session cache")
	flag.Int64Var(&clConf.DataCacheSize, "datacache", defConf.DataCacheSize, "integer count of items in the data cache")
	flgOverride := flag.Bool("override", false, "override any congfig file with command line arguments")

	flag.Parse()

	clConf.Ports = clPorts

	clConf.Seeds = []string{*flgDnsSeed}

	return &clConf, *flgOverride
}

func loadConf() error {

	cl, override := parseFlags()

	createConfig := false

	var err error
	// got no config file - make one
	if confFilePath == "" {
		log.Println("Debug - creating conf from command line " + cl.Ports.Bleeter)
		createConfig = true

		if cl.StoreRoot == defConf.StoreRoot { // set empty to find users home
			cl.StoreRoot = ""
		}

		setRootPath(cl.StoreRoot, cl.Ports.Bleeter)

		// use the command line config
		config = *cl

		// update the dataroot with the computed root
		config.StoreRoot = storeRoot

		// default config file location
		err = os.MkdirAll(storeRoot+"conf/", 0750)
		if err != nil {
			log.Fatal("Error - cant make conf directory")
		}
		confFilePath = storeRoot + "conf/config.json"
	} else {
		log.Println("Debug - loading conf file " + confFilePath)
		var body []byte
		body, err = ioutil.ReadFile(confFilePath)
		if err == nil {
			err = json.Unmarshal(body, &config)
		}
	}

	if err != nil {
		log.Fatal("Error - " + err.Error())
	}
	// if we didnt get a config file or if override
	if err != nil || override {
		config = *cl // use the command line config
	}

	println(config.Ports.Bleeter)

	// if we didnt have an existing config - of if we asked to recreate with the overrides..
	if createConfig || override {
		log.Println("Debug - creating conf at " + confFilePath)
		err = SaveConf()
		if err != nil {
			log.Println("Error - saving config - " + err.Error())
		}
	}

	forceFolders()

	return nil
}

func forceFolders() {
	// tidy it up and enforce trailing space
	storeRoot = filepath.Clean(storeRoot) + "/"

	if storeRoot == "//" {
		log.Fatal("Fatal - You cant use the root folder for Grex data")
	}

	if storeRoot == "" {
		log.Fatal("Fatal - You cant use '' for Grex data")
	}

	log.Println("Debug - store root - " + storeRoot)
	// ensure the data root folder exists
	dataRoot = storeRoot + "data/"

	log.Println("Debug - data store - " + dataRoot)

	// make sure the data root is in place
	err := os.MkdirAll(dataRoot, 0750)
	if err != nil {
		log.Fatal("Error - cant make root directory -" + err.Error())
	}
}

func SaveConf() error {

	type Animal struct {
		Name  string
		Order string
	}

	configBlob, err := json.MarshalIndent(&config, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(confFilePath, configBlob, 0640)
	if err != nil {
		return err
	}

	return nil
}

// expose the config settings for any potential app server
func GetAppServerAddress() string {
	return config.ExternalName + ":" + config.Ports.ExtRest
}
