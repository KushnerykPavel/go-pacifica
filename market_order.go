package pacifica

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateMarketOrderRequest represents the request data for creating a market order
type CreateMarketOrderRequest struct {
	Symbol          string    `json:"symbol"`
	Amount          string    `json:"amount"`
	Side            OrderSide `json:"side"`
	SlippagePercent string    `json:"slippage_percent"`
	ReduceOnly      bool      `json:"reduce_only"`
	ClientOrderID   string    `json:"client_order_id"`
	TakeProfit      *Target   `json:"take_profit,omitempty"`
	StopLoss        *Target   `json:"stop_loss,omitempty"`
	ExpiryWindow    int       `json:"expiry_window,omitempty"`
}

func (r CreateMarketOrderRequest) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// CreateMarketOrderOptions contains optional parameters for creating a market order
type CreateMarketOrderOptions struct {
	ClientOrderID string
	AgentWallet   *string
	ExpiryWindow  int64
}

// BuildCreateMarketOrderRequest builds a signed request for creating a market order
func (s *Exchange) BuildCreateMarketOrderRequest(params CreateMarketOrderRequest, opts *CreateMarketOrderOptions) (map[string]interface{}, error) {
	// Validate required fields
	if params.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if params.Amount == "" {
		return nil, fmt.Errorf("amount is required")
	}
	if params.Side != SideBid && params.Side != SideAsk {
		return nil, fmt.Errorf("side must be 'bid' or 'ask'")
	}
	if params.SlippagePercent == "" {
		return nil, fmt.Errorf("slippage_percent is required")
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var operationData map[string]interface{}
	if err := json.Unmarshal(data, &operationData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// Determine expiry window
	expiryWindow := int64(0)
	if opts != nil && opts.ExpiryWindow != 0 {
		expiryWindow = opts.ExpiryWindow
	}

	// Build signed request with operation type "create_market_order"
	request, err := s.BuildSignedRequest("create_market_order", operationData, expiryWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Add agent_wallet if provided
	if opts != nil && opts.AgentWallet != nil {
		request["agent_wallet"] = *opts.AgentWallet
	}

	return request, nil
}

// CreateMarketOrderResponse represents the response from the create market order endpoint
type CreateMarketOrderResponse struct {
	OrderID int64 `json:"order_id"`
}

// CreateMarketOrderError represents an error response from the API
type CreateMarketOrderError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// CreateMarketOrder creates a market order on Pacifica
func (c *RESTClient) CreateMarketOrder(params CreateMarketOrderRequest, opts *CreateMarketOrderOptions) (*CreateMarketOrderResponse, error) {
	// Build signed request
	request, err := c.signer.BuildCreateMarketOrderRequest(params, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/orders/create_market", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		var response CreateMarketOrderResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &response, nil
	case http.StatusBadRequest:
		var apiError CreateMarketOrderError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("bad request: %s", string(body))
		}
		return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
	default:
		var apiError CreateMarketOrderError
		if err := json.Unmarshal(body, &apiError); err == nil {
			return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}
