package src

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
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

func (rp *RewriteProxy) rewriteBody(resp *http.Response, prefix string) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	bodyStr := string(body)

	replacements := []struct{ old, new string }{
		{`src="/`, `src="` + prefix + `/`},
		{`href="/`, `href="` + prefix + `/`},
		{`action="/`, `action="` + prefix + `/`},
		{`"/assets/`, `"` + prefix + `/assets/`},
		{`"/api/`, `"` + prefix + `/api/`},
		{`"/ws`, `"` + prefix + `/ws`},
		{`"/socket`, `"` + prefix + `/socket`},
		{`" /site.webmanifest`, `" ` + prefix + `/site.webmanifest`},
	}

	for _, rp := range replacements {
		bodyStr = strings.ReplaceAll(bodyStr, rp.old, rp.new)
	}

	bodyStr = strings.ReplaceAll(bodyStr, prefix+`//`, prefix+`/`)

	resp.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
	resp.ContentLength = int64(len(bodyStr))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyStr)))
}
