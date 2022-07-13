/*
用于节点间的通信
*/
package ccache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HTTPPool ...
type HTTPPool struct {
	self string
	// 前缀路径
	basePath string
}

const defaultBasePath = "/ccache/"

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.ByteSlice())

}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}
