package httpclient

import (
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/auditr-io/testmock"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_ReusesTransport(t *testing.T) {
	var wg sync.WaitGroup
	expectedClients := 10
	wg.Add(expectedClients)

	clients := make([]*http.Client, expectedClients)
	for i := 0; i < expectedClients; i++ {
		go func(n int) {
			defer wg.Done()
			client, err := NewClient("https://auditr.io", nil, nil)
			assert.NoError(t, err)
			clients[n] = client
		}(i)
	}

	wg.Wait()
	assert.Equal(t, expectedClients, len(clients))
	for i, c := range clients {
		assert.NotNil(t, clients[i])
		assert.Equal(t, clients[0].Transport, c.Transport)
	}
}

func TestNewClient_WithSettings(t *testing.T) {
	var wg sync.WaitGroup
	expectedClients := 10
	wg.Add(expectedClients)

	tr, err := NewTransport(&HTTPClientSettings{
		Connect:          2 * time.Second,
		ExpectContinue:   1 * time.Second,
		IdleConn:         90 * time.Second,
		ConnKeepAlive:    30 * time.Second,
		MaxAllIdleConns:  100,
		MaxHostIdleConns: runtime.GOMAXPROCS(0) + 1,
		ResponseHeader:   2 * time.Second,
		TLSHandshake:     2 * time.Second,
	})
	assert.NoError(t, err)

	clients := make([]*http.Client, expectedClients)
	for i := 0; i < expectedClients; i++ {
		go func(n int) {
			defer wg.Done()
			client, err := NewClient(
				"https://auditr.io",
				tr,
				nil,
			)
			assert.NoError(t, err)
			clients[n] = client
		}(i)
	}

	wg.Wait()
	assert.Equal(t, expectedClients, len(clients))
	for i, c := range clients {
		assert.NotNil(t, clients[i])
		assert.Equal(t, clients[0].Transport, c.Transport)
	}
}

func TestNewClient_WithHeader(t *testing.T) {
	expectedHeader := http.Header{
		"Authorization": []string{
			"Bearer xxx",
		},
	}

	m := &testmock.MockTransport{
		RoundTripFn: func(m *testmock.MockTransport, req *http.Request) (*http.Response, error) {
			assert.Equal(t, expectedHeader["Authorization"], req.Header["Authorization"])
			return &http.Response{
				StatusCode: 200,
			}, nil
		},
	}

	url := "https://auditr.io"
	client, err := NewClient(
		url,
		m,
		expectedHeader,
	)
	assert.NoError(t, err)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	client.Do(req)
}
