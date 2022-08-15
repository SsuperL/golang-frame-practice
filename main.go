package main

import (
	"ccache"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sorm"
	"sorm/logger"
	"surpc/xclient"
	"sync"
	"time"

	"surpc"

	_ "github.com/mattn/go-sqlite3"
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

func ccacheRun() {
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

func ormRun() {
	driver, source := "sqlite3", "sorm.db"
	engine, err := sorm.NewEngine(driver, source)
	defer engine.Close()

	if err != nil {
		logger.Error("failed to create engine")
	}
	session := engine.NewSession()
	session.DB().Exec("CREATE TABLE IF NOT EXISTS users(name text,age int);")
	session.DB().Exec("INSERT INTO users VALUES(?,?),(?,?);", "John", 13, "Amy", 15)
	rows, err := session.DB().Query("SELECT * FROM users;")
	if err != nil {
		logger.Error(err.Error())
	}
	fmt.Println(rows)
}

func startServer(addr chan string) {
	var foo Foo
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("Failed to listen err:", err)
	}

	server := surpc.NewServer()
	_ = server.Register(&foo)
	log.Println("listening on ", lis.Addr().String())
	// surpc.HandleHTTP()
	addr <- lis.Addr().String()
	// http.Serve(lis, nil)

	server.Accept(lis)
}

type Foo struct{}
type Args struct{ Arg1, Arg2 int }

func (f Foo) Sum(args Args, res *int) error {
	*res = args.Arg1 + args.Arg2
	return nil
}

func (f Foo) Sleep(args Args, res *int) error {
	time.Sleep(time.Second * time.Duration(args.Arg1))
	*res = args.Arg1 + args.Arg2
	return nil
}

func foo(ctx context.Context, xc *xclient.XClient, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Arg1, args.Arg2, reply)
	}

}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, nil, xclient.RandomSelect)
	defer func() {
		_ = xc.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(context.Background(), xc, "call", "Foo.Sum", &Args{Arg1: i, Arg2: i * i})
		}(i)
	}
	wg.Wait()
}
func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, nil, xclient.RandomSelect)
	defer func() {
		_ = xc.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(context.Background(), xc, "broadcast", "Foo.Sum", &Args{Arg1: i, Arg2: i * i})
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(ctx, xc, "broadcast", "Foo.Sum", &Args{Arg1: i, Arg2: i * i})
		}(i)
	}
	wg.Wait()
}
func rpcRun(addr chan string) {
	address := <-addr
	client, _ := surpc.DialHTTP("tcp", address)
	fmt.Printf("client: %#v \n", client)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// _ = json.NewEncoder(conn).Encode(surpc.DefaultOption)
	// cc := codec.NewGobCodec(conn)
	// for i := 0; i < 5; i++ {
	// 	h := &codec.Header{
	// 		ServiceMethod: "test",
	// 		Seq:           i,
	// 	}
	// 	_ = cc.Write(h, fmt.Sprintf("rpc request %d", h.Seq))
	// 	_ = cc.ReadHeader(h)
	// 	log.Println("response h :", h)
	// 	var reply string
	// 	_ = cc.ReadBody(&reply)
	// 	log.Println("reply:", reply)
	// }
	// time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	// for i := 1; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		args := &Args{Arg1: i, Arg2: i * i}
	// 		var res int
	// 		// ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	// 		err := client.Call(context.Background(), "Foo.Sum", args, &res)
	// 		if err != nil {
	// 			log.Fatal("call test error: ", err)
	// 		}
	// 		log.Printf("reply: %d + %d = %d", i, i*i, res)

	// 	}(i)
	// }
	wg.Add(1)
	go func() {
		defer wg.Done()
		args := &Args{Arg1: 1, Arg2: 1}
		var res int
		err := client.Call(context.Background(), "Foo.Sum", args, &res)
		if err != nil {
			log.Fatal("call test error: ", err)
		}
		log.Printf("reply: %d + %d = %d", 1, 1, res)
	}()
	wg.Wait()
	// time.Sleep(2 * time.Second)
}

func main() {
	// ccacheRun()
	log.SetFlags(0)
	// ch := make(chan string)
	// go rpcRun(ch)
	ch1 := make(chan string)
	ch2 := make(chan string)

	go startServer(ch1)
	go startServer(ch2)
	addr1 := <-ch1
	addr2 := <-ch2
	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)

}
