package httpclient

import (
	"net/http"
	neturl "net/url"
	"runtime"
	"sync"
	"time"
)

var (
	// DefaultHTTPClientSettings contains reasonable default settings.
	DefaultHTTPClientSettings = &HTTPClientSettings{
		Connect:          2 * time.Second,
		ExpectContinue:   1 * time.Second,
		IdleConn:         90 * time.Second,
		ConnKeepAlive:    30 * time.Second,
		MaxAllIdleConns:  100,
		MaxHostIdleConns: runtime.GOMAXPROCS(0) + 1,
		ResponseHeader:   4 * time.Second,
		TLSHandshake:     4 * time.Second,
	}

	transports     = make(map[string]http.RoundTripper)
	transportsSync sync.Mutex
)

// HTTPClientSettings defines the HTTP setting for clients
type HTTPClientSettings struct {
	Connect          time.Duration
	ConnKeepAlive    time.Duration
	ExpectContinue   time.Duration
	IdleConn         time.Duration
	MaxAllIdleConns  int
	MaxHostIdleConns int
	ResponseHeader   time.Duration
	TLSHandshake     time.Duration
}

// NewClient creates an HTTP client with custom settings and headers.
func NewClient(
	url string,
	transport http.RoundTripper,
	headers http.Header,
) (*http.Client, error) {
	transportsSync.Lock()
	defer transportsSync.Unlock()

	var err error
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	tr, ok := transports[u.Host]
	if !ok {
		if transport != nil {
			tr = transport
		} else {
			tr, err = NewTransport(DefaultHTTPClientSettings)
		}

		transports[u.Host] = tr
	}

	client := &http.Client{
		Transport: &transportWrapper{
			Base:    tr,
			Headers: headers,
		},
	}

	return client, err
}
