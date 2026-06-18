package src

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

type RewriteProxy struct {
	target   string
	endpoint string
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
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = pr.In.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			ct := resp.Header.Get("Content-Type")
			if strings.Contains(ct, "text/html") {
				rp.rewriteBody(resp, "/"+endpoint)
			}
			return nil
		},
	}

	rp.proxy = proxy
	return rp, nil
}

func (rp *RewriteProxy) Handler() http.Handler {
	return rp.proxy
}

var (
	pathAttrRE   = regexp.MustCompile(`((?:src|href|action|content|data-src|to)=["'])/([^"'\s>]*["'])`)
	inlinePathRE = regexp.MustCompile(`(["'\(\s])/((?:assets|api|ws|socket|_next|static|site)[^"'\s,\)\]]*)`)
)

func (rp *RewriteProxy) rewriteBody(resp *http.Response, prefix string) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	bodyStr := string(body)

	if strings.Contains(bodyStr, prefix+"/"+prefix) {
		return
	}

	bodyStr = pathAttrRE.ReplaceAllString(bodyStr, `${1}`+prefix+`/$2`)
	bodyStr = inlinePathRE.ReplaceAllString(bodyStr, `${1}`+prefix+`/$2`)

	resp.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
	resp.ContentLength = int64(len(bodyStr))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyStr)))
}
