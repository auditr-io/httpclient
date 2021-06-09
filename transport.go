package httpclient

import (
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

// NewTransport creates a new transport with given settings.
// If settings is nil, the DefaultHTTPClientSettings will be used.
func NewTransport(settings *HTTPClientSettings) (*http.Transport, error) {
	if settings == nil {
		settings = DefaultHTTPClientSettings
	}

	tr := &http.Transport{
		ResponseHeaderTimeout: settings.ResponseHeader,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: settings.ConnKeepAlive,
			Timeout:   settings.Connect,
		}).DialContext,
		MaxIdleConns:          settings.MaxAllIdleConns,
		IdleConnTimeout:       settings.IdleConn,
		TLSHandshakeTimeout:   settings.TLSHandshake,
		MaxIdleConnsPerHost:   settings.MaxHostIdleConns,
		ExpectContinueTimeout: settings.ExpectContinue,
		ForceAttemptHTTP2:     true,
	}

	// So client makes HTTP/2 requests
	err := http2.ConfigureTransport(tr)

	return tr, err
}

// transportWrapper is the wrapper transport for HTTP client
type transportWrapper struct {
	Base    http.RoundTripper
	Headers http.Header
}

// RoundTrip appends additional headers to all requests
func (t *transportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				req.Body.Close()
			}
		}()
	}

	req2 := cloneRequest(req)
	// copy additional headers
	for k, s := range t.Headers {
		req2.Header[k] = append([]string(nil), s...)
	}

	reqBodyClosed = true
	return t.Base.RoundTrip(req2)
}

// cloneRequest clones the request for modification
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
