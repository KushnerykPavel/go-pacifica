package pacifica

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OrderSide represents the order side (bid or ask)
type OrderSide string

const (
	SideBid OrderSide = "bid"
	SideAsk OrderSide = "ask"
)

// TimeInForce represents the time in force for an order
type TimeInForce string

const (
	TIFGTC TimeInForce = "GTC" // Good Till Cancel
	TIFIOC TimeInForce = "IOC" // Immediate Or Cancel
	TIFALO TimeInForce = "ALO" // Allow Override
)

// Target represents take profit stop order configuration
type Target struct {
	StopPrice     string `json:"stop_price"`
	LimitPrice    string `json:"limit_price,omitempty"`
	ClientOrderID string `json:"client_order_id,omitempty"`
}

// CreateLimitOrderRequest represents the request data for creating a limit order
type CreateLimitOrderRequest struct {
	Symbol        string      `json:"symbol"`
	Price         string      `json:"price"`
	Amount        string      `json:"amount"`
	Side          OrderSide   `json:"side"`
	TIF           TimeInForce `json:"tif"`
	ReduceOnly    bool        `json:"reduce_only"`
	ClientOrderID string      `json:"client_order_id"`
	TakeProfit    *Target     `json:"take_profit,omitempty"`
	StopLoss      *Target     `json:"stop_loss,omitempty"`
	ExpiryWindow  int         `json:"expiry_window,omitempty"`
}

func (r CreateLimitOrderRequest) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// CreateLimitOrderOptions contains optional parameters for creating a limit order
type CreateLimitOrderOptions struct {
	ClientOrderID string
	AgentWallet   *string
	ExpiryWindow  int64
}

// BuildCreateLimitOrderRequest builds a signed request for creating a limit order
func (s *Exchange) BuildCreateLimitOrderRequest(params CreateLimitOrderRequest, opts *CreateLimitOrderOptions) (map[string]interface{}, error) {
	// Validate required fields
	if params.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if params.Price == "" {
		return nil, fmt.Errorf("price is required")
	}
	if params.Amount == "" {
		return nil, fmt.Errorf("amount is required")
	}
	if params.Side != SideBid && params.Side != SideAsk {
		return nil, fmt.Errorf("side must be 'bid' or 'ask'")
	}
	if params.TIF != TIFGTC && params.TIF != TIFIOC && params.TIF != TIFALO {
		return nil, fmt.Errorf("tif must be 'GTC', 'IOC', or 'ALO'")
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

	// Build signed request
	request, err := s.BuildSignedRequest("create_order", operationData, expiryWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Add agent_wallet if provided
	if opts != nil && opts.AgentWallet != nil {
		request["agent_wallet"] = *opts.AgentWallet
	}

	return request, nil
}

// CreateLimitOrderResponse represents the response from the create limit order endpoint
type CreateLimitOrderResponse struct {
	OrderID int64 `json:"order_id"`
}

// CreateLimitOrderError represents an error response from the API
type CreateLimitOrderError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// CreateLimitOrder creates a limit order on Pacifica
func (c *RESTClient) CreateLimitOrder(params CreateLimitOrderRequest, opts *CreateLimitOrderOptions) (*CreateLimitOrderResponse, error) {
	// Build signed request
	request, err := c.signer.BuildCreateLimitOrderRequest(params, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/orders/create", bytes.NewBuffer(jsonData))
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
		var response CreateLimitOrderResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &response, nil
	case http.StatusBadRequest:
		var apiError CreateLimitOrderError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("bad request: %s", string(body))
		}
		return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
	default:
		var apiError CreateLimitOrderError
		if err := json.Unmarshal(body, &apiError); err == nil {
			return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}
