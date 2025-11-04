package pacifica

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CancelOrderRequest represents the request data for canceling an order
type CancelOrderRequest struct {
	Symbol        string `json:"symbol"`
	OrderID       *int64 `json:"order_id,omitempty"`
	ClientOrderID string `json:"client_order_id,omitempty"`
}

func (r CancelOrderRequest) String() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// CancelOrderOptions contains optional parameters for canceling an order
type CancelOrderOptions struct {
	AgentWallet  *string
	ExpiryWindow int64
}

// BuildCancelOrderRequest builds a signed request for canceling an order
func (s *Exchange) BuildCancelOrderRequest(params CancelOrderRequest, opts *CancelOrderOptions) (map[string]interface{}, error) {
	// Validate required fields
	if params.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	// Either order_id or client_order_id must be provided
	if params.OrderID == nil && params.ClientOrderID == "" {
		return nil, fmt.Errorf("either order_id or client_order_id is required")
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var operationData map[string]interface{}
	if err := json.Unmarshal(data, &operationData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	// Remove empty fields from operation data
	if operationData["order_id"] == nil {
		delete(operationData, "order_id")
	}
	if operationData["client_order_id"] == "" {
		delete(operationData, "client_order_id")
	}

	// Determine expiry window
	expiryWindow := int64(0)
	if opts != nil && opts.ExpiryWindow != 0 {
		expiryWindow = opts.ExpiryWindow
	}

	// Build signed request with operation type "cancel_order"
	request, err := s.BuildSignedRequest("cancel_order", operationData, expiryWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Add agent_wallet if provided
	if opts != nil && opts.AgentWallet != nil {
		request["agent_wallet"] = *opts.AgentWallet
	}

	return request, nil
}

// CancelOrderResponse represents the response from the cancel order endpoint
type CancelOrderResponse struct {
	Success bool `json:"success"`
}

// CancelOrderError represents an error response from the API
type CancelOrderError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// CancelOrder cancels an order on Pacifica
func (c *RESTClient) CancelOrder(params CancelOrderRequest, opts *CancelOrderOptions) (*CancelOrderResponse, error) {
	// Build signed request
	request, err := c.signer.BuildCancelOrderRequest(params, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build signed request: %w", err)
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/orders/cancel", bytes.NewBuffer(jsonData))
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
		var response CancelOrderResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &response, nil
	case http.StatusBadRequest:
		var apiError CancelOrderError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("bad request: %s", string(body))
		}
		return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
	default:
		var apiError CancelOrderError
		if err := json.Unmarshal(body, &apiError); err == nil {
			return nil, fmt.Errorf("API error (code %d): %s", apiError.Code, apiError.Error)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}
