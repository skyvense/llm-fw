package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy 封装了反向代理的功能
type ReverseProxy struct {
	TargetURL string
	proxy     *httputil.ReverseProxy
}

// NewReverseProxy 创建一个新的反向代理
func NewReverseProxy(targetURL string) (*ReverseProxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &ReverseProxy{
		TargetURL: targetURL,
		proxy:     httputil.NewSingleHostReverseProxy(url),
	}, nil
}

// ServeHTTP 实现了 http.Handler 接口
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
