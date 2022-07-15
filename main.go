package main

import (
	"ccache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"A": "A",
	"B": "B",
	"C": "C",
}

func createGroup() *ccache.Group {
	return ccache.NewGroup("source", 2<<10, ccache.GetterFunc(func(key string) ([]byte, error) {
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return []byte(fmt.Sprintf("%s not exists \n", key)), nil
	}))
}

func startAPIServer(apiAddr string, group *ccache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(value.ByteSlice())
	}))
	log.Println("fontend server is running at:", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr, nil))
}

func main() {
	var api bool
	var port int

	flag.IntVar(&port, "port", 8001, "Ccache server port")
	flag.BoolVar(&api, "api", false, "Start an api server")
	flag.Parse()

	group := createGroup()
	addr := ":9090"
	peerList := map[int]string{
		8001: "localhost:8081",
		8002: "localhost:8082",
		8003: "localhost:8083"}

	peers := ccache.NewHTTPPoolWithOpts(peerList[port], ccache.HTTPPoolOptions{})
	var addrs []string
	for _, addr := range peerList {
		addrs = append(addrs, addr)
	}
	peers.Set(addrs...)
	group.RegisterPeers(peers)

	if api {
		go startAPIServer(addr, group)
	}
	log.Println("ccache is running at: ", peerList[port])
	log.Fatal(http.ListenAndServe(peerList[port], peers))
}
