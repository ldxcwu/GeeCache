package geecache

import (
	"net/http"
	"strings"
)

type HttpServer struct {
	addr     string
	basePath string
}

func NewHTTPServer(addr, basePath string) *HttpServer {
	return &HttpServer{
		addr:     addr,
		basePath: basePath,
	}
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, s.basePath) {
		http.Error(w, "Invalid request URL(without cacheGroup):"+req.URL.Path, http.StatusBadRequest)
		return
	}
	parts := strings.SplitN(req.URL.Path[len(s.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid request URL(invalid format:cacheGroupName/key):"+req.URL.Path, http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "No such cache group: "+groupName, http.StatusNotFound)
		return
	}
	data, err := group.Get(key)
	if err != nil {
		http.Error(w, "something went wrong:"+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data.ByteSlice())
}
