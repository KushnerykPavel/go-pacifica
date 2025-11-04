package pacifica

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCancelOrderRequest(t *testing.T) {
	signer := generateTestExchange(t)

	tests := []struct {
		name     string
		params   CancelOrderRequest
		opts     *CancelOrderOptions
		wantErr  bool
		validate func(*testing.T, map[string]interface{})
	}{
		{
			name: "cancel order with order_id",
			params: CancelOrderRequest{
				Symbol:  "BTC",
				OrderID: intPtr(12345),
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "BTC", req["symbol"])
				// JSON unmarshaling converts numbers to float64
				orderID, ok := req["order_id"].(float64)
				require.True(t, ok)
				assert.Equal(t, float64(12345), orderID)
				assert.NotContains(t, req, "client_order_id")
				assert.Contains(t, req, "account")
				assert.Contains(t, req, "signature")
				assert.Contains(t, req, "timestamp")
			},
		},
		{
			name: "cancel order with client_order_id",
			params: CancelOrderRequest{
				Symbol:        "ETH",
				ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "ETH", req["symbol"])
				assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
				assert.NotContains(t, req, "order_id")
				assert.Contains(t, req, "account")
				assert.Contains(t, req, "signature")
				assert.Contains(t, req, "timestamp")
			},
		},
		{
			name: "cancel order with both order_id and client_order_id",
			params: CancelOrderRequest{
				Symbol:        "BTC",
				OrderID:       intPtr(12345),
				ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "BTC", req["symbol"])
				// JSON unmarshaling converts numbers to float64
				orderID, ok := req["order_id"].(float64)
				require.True(t, ok)
				assert.Equal(t, float64(12345), orderID)
				assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
			},
		},
		{
			name: "cancel order with agent_wallet",
			params: CancelOrderRequest{
				Symbol:  "BTC",
				OrderID: intPtr(12345),
			},
		opts: &CancelOrderOptions{
			AgentWallet: func() *string { s := "69trU9A5..."; return &s }(),
		},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "69trU9A5...", req["agent_wallet"])
			},
		},
		{
			name: "cancel order with custom expiry_window",
			params: CancelOrderRequest{
				Symbol:  "BTC",
				OrderID: intPtr(12345),
			},
			opts: &CancelOrderOptions{
				ExpiryWindow: 10000,
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, int64(10000), req["expiry_window"])
			},
		},
		{
			name: "missing symbol",
			params: CancelOrderRequest{
				OrderID: intPtr(12345),
			},
			wantErr: true,
		},
		{
			name: "missing both order_id and client_order_id",
			params: CancelOrderRequest{
				Symbol: "BTC",
			},
			wantErr: true,
		},
		{
			name: "empty client_order_id",
			params: CancelOrderRequest{
				Symbol:        "BTC",
				ClientOrderID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := signer.BuildCancelOrderRequest(tt.params, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, req)
			}
		})
	}
}

func TestCancelOrderRequestFromDocumentation(t *testing.T) {
	// Test the exact example from the Pacifica documentation
	// https://docs.pacifica.fi/api-documentation/api/rest-api/orders/cancel-order

	signer := generateTestExchange(t)

	params := CancelOrderRequest{
		Symbol:  "BTC",
		OrderID: intPtr(123),
	}

	opts := &CancelOrderOptions{
		AgentWallet:  func() *string { s := "69trU9A5..."; return &s }(),
		ExpiryWindow: 30000,
	}

	req, err := signer.BuildCancelOrderRequest(params, opts)
	require.NoError(t, err)

	// Verify all fields match documentation
	assert.Equal(t, "BTC", req["symbol"])
	// JSON unmarshaling converts numbers to float64
	orderID, ok := req["order_id"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(123), orderID)
	assert.Equal(t, "69trU9A5...", req["agent_wallet"])
	assert.Equal(t, int64(30000), req["expiry_window"])

	// Verify authentication fields
	assert.Contains(t, req, "account")
	assert.Contains(t, req, "signature")
	assert.Contains(t, req, "timestamp")
}

func TestCancelOrderRequestWithClientOrderID(t *testing.T) {
	signer := generateTestExchange(t)

	params := CancelOrderRequest{
		Symbol:        "BTC",
		ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
	}

	req, err := signer.BuildCancelOrderRequest(params, nil)
	require.NoError(t, err)

	// Verify client_order_id is used instead of order_id
	assert.Equal(t, "BTC", req["symbol"])
	assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
	assert.NotContains(t, req, "order_id")
}

func TestRESTClientCancelOrder(t *testing.T) {
	signer := generateTestExchange(t)

	client := NewRESTClient("", signer)
	assert.NotNil(t, client)
	assert.Equal(t, MainnetAPIURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, signer, client.signer)
}

// Helper function for tests
func intPtr(i int64) *int64 {
	return &i
}
