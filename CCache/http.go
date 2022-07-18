/*
用于节点间的通信
*/
package ccache

import (
	"ccache/ccachepb"
	"ccache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

// HTTPPool ...
type HTTPPool struct {
	self string
	// 前缀路径
	basePath    string
	mu          sync.Mutex             // guards
	peers       *consistenthash.Map    //节点列表
	httpGetters map[string]*httpGetter //映射节点和路径关系（baseURL前缀）
	opts        HTTPPoolOptions
}

// HTTP客户端
type httpGetter struct {
	baseURL string // e.g http://localhost:8080
}

const (
	defaultBasePath = "/ccache/"
	defaultReplicas = 3
)

type HTTPPoolOptions struct {
	replicas int
}

func NewHTTPPoolWithOpts(self string, opts HTTPPoolOptions) *HTTPPool {
	hp := &HTTPPool{
		self:        self,
		basePath:    defaultBasePath,
		httpGetters: make(map[string]*httpGetter),
	}
	if opts.replicas == 0 {
		hp.opts.replicas = defaultReplicas
	}

	return hp
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, defaultBasePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// 假设请求路径是 /<basePath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	groupname := parts[0]
	key := parts[1]

	group := GetGroup(groupname)
	if group == nil {
		http.Error(w, "No such group", http.StatusNotFound)
		return
	}

	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response, err := proto.Marshal(&ccachepb.Response{Value: value.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(response)

}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (h *httpGetter) Get(req *ccachepb.Request) (response *ccachepb.Response, err error) {
	url := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(req.GetGroup()), url.QueryEscape(req.GetKey()))
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server response status:%v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body:%v", err)
	}

	err = proto.Unmarshal(bytes, response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to proto error: %v", err)
	}
	return
}

var _ PeerGetter = (*httpGetter)(nil)

// Set 更新远程节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.NewMap(p.opts.replicas, nil)
	// 映射节点和getter关系
	p.peers.Add(peers...)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer pick a peer
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.peers == nil {
		return nil, false
	}
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		return p.httpGetters[peer], true
	}

	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
