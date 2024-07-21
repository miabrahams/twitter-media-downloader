package lib

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func NewHTTPClient(proxyURL string) HTTPClient {
	// Create and return an http.Client with the given proxy settings

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Duration(5) * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Duration(5) * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
		},
	}

	return client
}

func NewProxyClient(proxyURL string) HTTPClient {
	parsedURL, _ := url.Parse(proxyURL)
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		},
	}
}
