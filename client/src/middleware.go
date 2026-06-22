package src

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type RewriteProxy struct {
	target   string
	endpoint string
	prefix   string
	proxy    *httputil.ReverseProxy
}

func NewRewriteProxy(targetURL, endpoint string) (*RewriteProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	rp := &RewriteProxy{
		target:   targetURL,
		endpoint: endpoint,
		prefix:   "/" + endpoint,
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = pr.In.Host
			if strings.HasPrefix(pr.Out.URL.Path, rp.prefix+"/") {
				pr.Out.URL.Path = strings.TrimPrefix(pr.Out.URL.Path, rp.prefix)
			} else if pr.Out.URL.Path == rp.prefix {
				pr.Out.URL.Path = "/"
			}
			if pr.Out.URL.RawPath != "" {
				pr.Out.URL.RawPath = ""
			}
		},
	}

	rp.proxy = proxy
	return rp, nil
}

func (rp *RewriteProxy) Handler() http.Handler {
	return rp.proxy
}
