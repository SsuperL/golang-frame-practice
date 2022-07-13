package main

import (
	"ccache"
	"log"
	"net/http"
)

var db = map[string]string{
	"A": "A",
	"B": "B",
	"C": "C",
}

func main() {
	ccache.NewGroup("source", 2<<10, ccache.GetterFunc(func(key string) ([]byte, error) {
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, nil
	}))

	addr := ":9090"
	peers := ccache.NewHTTPPool(addr)
	log.Println("ccache is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
