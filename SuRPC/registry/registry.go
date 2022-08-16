// Package registry ...
// 服务注册
package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type SuRegistry struct {
	// 超时时间，默认5min
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_surpc_/registry"
	defaultTimeout = time.Minute * 5
)

// NewSuRegistry init
func NewSuRegistry(timeout time.Duration) *SuRegistry {
	return &SuRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultRegistry = NewSuRegistry(defaultTimeout)

// putServer 如果存在列表中则更新start，不存在则新建
func (s *SuRegistry) putServer(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	server := s.servers[addr]
	if server == nil {
		s.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.servers[addr].start = time.Now()
	}
}

// 返回可用的服务列表
func (s *SuRegistry) aliveServers() []string {
	var aliveServers []string
	for addr, server := range s.servers {
		if s.timeout == 0 || server.start.Add(s.timeout).After(time.Now()) {
			aliveServers = append(aliveServers, addr)
		} else {
			// 超时则移除服务
			delete(s.servers, addr)
		}
	}
	sort.Strings(aliveServers)
	return aliveServers
}

func (s *SuRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("X-Surpc-Servers", strings.Join(s.aliveServers(), ","))
	case "POST":
		addr := r.Header.Get("X-Surpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *SuRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, s)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultRegistry.HandleHTTP(defaultPath)
}

// Heartbeat send heartbeat
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}

	var err error
	err = sendHeartBeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartBeat(registry, addr)
		}
	}()
}

func sendHeartBeat(registry, addr string) error {
	log.Println(addr, "send heartbeat to registry ", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Surpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat: ", err)
		return err
	}
	return nil
}
