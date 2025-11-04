package pacifica

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateLimitOrderRequest(t *testing.T) {
	signer := generateTestExchange(t)

	tests := []struct {
		name     string
		params   CreateLimitOrderRequest
		opts     *CreateLimitOrderOptions
		wantErr  bool
		validate func(*testing.T, map[string]interface{})
	}{
		{
			name: "basic limit order",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "BTC", req["symbol"])
				assert.Equal(t, "50000", req["price"])
				assert.Equal(t, "0.1", req["amount"])
				assert.Equal(t, "bid", req["side"])
				assert.Equal(t, "GTC", req["tif"])
				assert.Equal(t, false, req["reduce_only"])
				assert.Contains(t, req, "account")
				assert.Contains(t, req, "signature")
				assert.Contains(t, req, "timestamp")
			},
		},
		{
			name: "order with client_order_id",
			params: CreateLimitOrderRequest{
				Symbol:        "ETH",
				Price:         "2000",
				Amount:        "1.5",
				Side:          SideAsk,
				TIF:           TIFIOC,
				ReduceOnly:    true,
				ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "ETH", req["symbol"])
				assert.Equal(t, "2000", req["price"])
				assert.Equal(t, "1.5", req["amount"])
				assert.Equal(t, "ask", req["side"])
				assert.Equal(t, "IOC", req["tif"])
				assert.Equal(t, true, req["reduce_only"])
				assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
			},
		},
		{
			name: "order with take_profit",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
				TakeProfit: &Target{
					StopPrice:     "55000",
					LimitPrice:    "54950",
					ClientOrderID: "e36ac10b-58cc-4372-a567-0e02b2c3d479",
				},
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Contains(t, req, "take_profit")
				takeProfit, ok := req["take_profit"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "55000", takeProfit["stop_price"])
				assert.Equal(t, "54950", takeProfit["limit_price"])
				assert.Equal(t, "e36ac10b-58cc-4372-a567-0e02b2c3d479", takeProfit["client_order_id"])
			},
		},
		{
			name: "order with stop_loss",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
				StopLoss: &Target{
					StopPrice:     "48000",
					LimitPrice:    "47950",
					ClientOrderID: "d25ac10b-58cc-4372-a567-0e02b2c3d479",
				},
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Contains(t, req, "stop_loss")
				stopLoss, ok := req["stop_loss"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "48000", stopLoss["stop_price"])
				assert.Equal(t, "47950", stopLoss["limit_price"])
				assert.Equal(t, "d25ac10b-58cc-4372-a567-0e02b2c3d479", stopLoss["client_order_id"])
			},
		},
		{
			name: "order with take_profit and stop_loss",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
				TakeProfit: &Target{
					StopPrice:  "55000",
					LimitPrice: "54950",
				},
				StopLoss: &Target{
					StopPrice:  "48000",
					LimitPrice: "47950",
				},
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Contains(t, req, "take_profit")
				assert.Contains(t, req, "stop_loss")
				takeProfit, ok := req["take_profit"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "55000", takeProfit["stop_price"])
				stopLoss, ok := req["stop_loss"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "48000", stopLoss["stop_price"])
			},
		},
		{
			name: "order with agent_wallet",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			opts: &CreateLimitOrderOptions{
				AgentWallet: stringPtr("69trU9A5..."),
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "69trU9A5...", req["agent_wallet"])
			},
		},
		{
			name: "order with custom expiry_window",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			opts: &CreateLimitOrderOptions{
				ExpiryWindow: 10000,
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, int64(10000), req["expiry_window"])
			},
		},
		{
			name: "missing symbol",
			params: CreateLimitOrderRequest{
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "missing price",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "missing amount",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "invalid side",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       OrderSide("invalid"),
				TIF:        TIFGTC,
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "invalid tif",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TimeInForce("invalid"),
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "order with take_profit without limit_price",
			params: CreateLimitOrderRequest{
				Symbol:     "BTC",
				Price:      "50000",
				Amount:     "0.1",
				Side:       SideBid,
				TIF:        TIFGTC,
				ReduceOnly: false,
				TakeProfit: &Target{
					StopPrice: "55000",
				},
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Contains(t, req, "take_profit")
				takeProfit, ok := req["take_profit"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "55000", takeProfit["stop_price"])
				assert.NotContains(t, takeProfit, "limit_price")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := signer.BuildCreateLimitOrderRequest(tt.params, tt.opts)
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

func TestCreateLimitOrderRequestFromDocumentation(t *testing.T) {
	// Test the exact example from the Pacifica documentation
	// https://docs.pacifica.fi/api-documentation/api/rest-api/orders/create-limit-order

	signer := generateTestExchange(t)

	params := CreateLimitOrderRequest{
		Symbol:        "BTC",
		Price:         "50000",
		Amount:        "0.1",
		Side:          SideBid,
		TIF:           TIFGTC,
		ReduceOnly:    false,
		ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
		TakeProfit: &Target{
			StopPrice:     "55000",
			LimitPrice:    "54950",
			ClientOrderID: "e36ac10b-58cc-4372-a567-0e02b2c3d479",
		},
		StopLoss: &Target{
			StopPrice:     "48000",
			LimitPrice:    "47950",
			ClientOrderID: "d25ac10b-58cc-4372-a567-0e02b2c3d479",
		},
	}

	opts := &CreateLimitOrderOptions{
		AgentWallet:  stringPtr("69trU9A5..."),
		ExpiryWindow: 30000,
	}

	req, err := signer.BuildCreateLimitOrderRequest(params, opts)
	require.NoError(t, err)

	// Verify all fields match documentation
	assert.Equal(t, "BTC", req["symbol"])
	assert.Equal(t, "50000", req["price"])
	assert.Equal(t, "0.1", req["amount"])
	assert.Equal(t, "bid", req["side"])
	assert.Equal(t, "GTC", req["tif"])
	assert.Equal(t, false, req["reduce_only"])
	assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
	assert.Equal(t, "69trU9A5...", req["agent_wallet"])
	assert.Equal(t, int64(30000), req["expiry_window"])

	// Verify take_profit
	takeProfit, ok := req["take_profit"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "55000", takeProfit["stop_price"])
	assert.Equal(t, "54950", takeProfit["limit_price"])
	assert.Equal(t, "e36ac10b-58cc-4372-a567-0e02b2c3d479", takeProfit["client_order_id"])

	// Verify stop_loss
	stopLoss, ok := req["stop_loss"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "48000", stopLoss["stop_price"])
	assert.Equal(t, "47950", stopLoss["limit_price"])
	assert.Equal(t, "d25ac10b-58cc-4372-a567-0e02b2c3d479", stopLoss["client_order_id"])

	// Verify authentication fields
	assert.Contains(t, req, "account")
	assert.Contains(t, req, "signature")
	assert.Contains(t, req, "timestamp")
}

func TestRESTClient(t *testing.T) {
	signer := generateTestExchange(t)

	client := NewRESTClient("", signer)
	assert.NotNil(t, client)
	assert.Equal(t, MainnetAPIURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, signer, client.signer)

	// Test with custom base URL
	customURL := "https://custom-api.example.com/api/v1"
	client2 := NewRESTClient(customURL, signer)
	assert.Equal(t, customURL, client2.baseURL)
}

func TestOrderConstants(t *testing.T) {
	// Test OrderSide constants
	assert.Equal(t, OrderSide("bid"), SideBid)
	assert.Equal(t, OrderSide("ask"), SideAsk)

	// Test TimeInForce constants
	assert.Equal(t, TimeInForce("GTC"), TIFGTC)
	assert.Equal(t, TimeInForce("IOC"), TIFIOC)
	assert.Equal(t, TimeInForce("ALO"), TIFALO)
}

func TestRESTClientMarketOrder(t *testing.T) {
	signer := generateTestExchange(t)

	client := NewRESTClient("", signer)
	assert.NotNil(t, client)
	assert.Equal(t, MainnetAPIURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, signer, client.signer)

	// Test with custom base URL
	customURL := "https://custom-api.example.com/api/v1"
	client2 := NewRESTClient(customURL, signer)
	assert.Equal(t, customURL, client2.baseURL)
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
