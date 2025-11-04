package pacifica

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateMarketOrderRequest(t *testing.T) {
	signer := generateTestExchange(t)

	tests := []struct {
		name     string
		params   CreateMarketOrderRequest
		opts     *CreateMarketOrderOptions
		wantErr  bool
		validate func(*testing.T, map[string]interface{})
	}{
		{
			name: "basic market order",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "BTC", req["symbol"])
				assert.Equal(t, "0.1", req["amount"])
				assert.Equal(t, "bid", req["side"])
				assert.Equal(t, "0.5", req["slippage_percent"])
				assert.Equal(t, false, req["reduce_only"])
				assert.Contains(t, req, "account")
				assert.Contains(t, req, "signature")
				assert.Contains(t, req, "timestamp")
			},
		},
		{
			name: "market order with client_order_id",
			params: CreateMarketOrderRequest{
				Symbol:          "ETH",
				Amount:          "1.5",
				Side:            SideAsk,
				SlippagePercent: "1.0",
				ReduceOnly:      true,
				ClientOrderID:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "ETH", req["symbol"])
				assert.Equal(t, "1.5", req["amount"])
				assert.Equal(t, "ask", req["side"])
				assert.Equal(t, "1.0", req["slippage_percent"])
				assert.Equal(t, true, req["reduce_only"])
				assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", req["client_order_id"])
			},
		},
		{
			name: "market order with take_profit",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
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
			name: "market order with stop_loss",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
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
			name: "market order with take_profit and stop_loss",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
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
			name: "market order with agent_wallet",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			opts: &CreateMarketOrderOptions{
				AgentWallet: stringPtr("69trU9A5..."),
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, "69trU9A5...", req["agent_wallet"])
			},
		},
		{
			name: "market order with custom expiry_window",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			opts: &CreateMarketOrderOptions{
				ExpiryWindow: 10000,
			},
			validate: func(t *testing.T, req map[string]interface{}) {
				assert.Equal(t, int64(10000), req["expiry_window"])
			},
		},
		{
			name: "missing symbol",
			params: CreateMarketOrderRequest{
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			wantErr: true,
		},
		{
			name: "missing amount",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			wantErr: true,
		},
		{
			name: "missing slippage_percent",
			params: CreateMarketOrderRequest{
				Symbol:     "BTC",
				Amount:     "0.1",
				Side:       SideBid,
				ReduceOnly: false,
			},
			wantErr: true,
		},
		{
			name: "invalid side",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            OrderSide("invalid"),
				SlippagePercent: "0.5",
				ReduceOnly:      false,
			},
			wantErr: true,
		},
		{
			name: "market order with take_profit without limit_price",
			params: CreateMarketOrderRequest{
				Symbol:          "BTC",
				Amount:          "0.1",
				Side:            SideBid,
				SlippagePercent: "0.5",
				ReduceOnly:      false,
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
			req, err := signer.BuildCreateMarketOrderRequest(tt.params, tt.opts)
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

func TestCreateMarketOrderRequestFromDocumentation(t *testing.T) {
	// Test the exact example from the Pacifica documentation
	// https://docs.pacifica.fi/api-documentation/api/rest-api/orders/create-market-order

	signer := generateTestExchange(t)

	params := CreateMarketOrderRequest{
		Symbol:          "BTC",
		Amount:          "0.1",
		Side:            SideBid,
		SlippagePercent: "0.5",
		ReduceOnly:      false,
		ClientOrderID:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
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

	opts := &CreateMarketOrderOptions{
		AgentWallet:  stringPtr("69trU9A5..."),
		ExpiryWindow: 30000,
	}

	req, err := signer.BuildCreateMarketOrderRequest(params, opts)
	require.NoError(t, err)

	// Verify all fields match documentation
	assert.Equal(t, "BTC", req["symbol"])
	assert.Equal(t, "0.1", req["amount"])
	assert.Equal(t, "bid", req["side"])
	assert.Equal(t, "0.5", req["slippage_percent"])
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
