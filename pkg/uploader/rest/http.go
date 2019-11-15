package rest

import (
	"context"
	"net/http"
	"net/url"
	"sync"
)

var (
	mu           sync.Mutex
	currentProxy string
	client       = new(http.Client)
)

func Do(ctx context.Context, proxy string, req *http.Request) (*http.Response, error) {
	mu.Lock()
	defer mu.Unlock()

	if currentProxy != proxy {
		if err := updateClient(proxy); err != nil {
			return nil, err
		}
	}
	return client.Do(req.WithContext(ctx))
}

func updateClient(newProxy string) error {
	if currentProxy == newProxy {
		return nil
	}

	client.CloseIdleConnections()

	if newProxy == "" {
		client = new(http.Client)
		return nil
	}

	proxyURL, err := url.Parse(newProxy)
	if err != nil {
		return err
	}
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	return nil
}
