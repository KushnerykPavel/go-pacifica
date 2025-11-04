package pacifica

import (
	"net/http"
	"time"
)

const (
	MainnetAPIURL = "https://api.pacifica.fi/api/v1"
	MainnetWSURL  = "wss://ws.pacifica.fi/ws"
)

// RESTClient handles REST API requests to Pacifica
type RESTClient struct {
	baseURL    string
	httpClient *http.Client
	signer     *Exchange
}

// NewRESTClient creates a new REST API client
func NewRESTClient(baseURL string, signer *Exchange) *RESTClient {
	if baseURL == "" {
		baseURL = MainnetAPIURL
	}
	return &RESTClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		signer: signer,
	}
}
