package pacifica_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KushnerykPavel/go-pacifica"
)

func TestWebsocketClient_Trades(t *testing.T) {
	client := pacifica.NewWebsocketClient(pacifica.MainnetWSURL)

	err := client.Connect(context.Background())
	assert.NoError(t, err)
	data := make(chan pacifica.Trades)
	errs := make(chan error)
	_, _ = client.Trades(pacifica.TradesSubscriptionParams{Symbol: "SOL"}, func(trades pacifica.Trades, err error) {
		if err != nil {
			errs <- err
			return
		}
		data <- trades
	})

	select {
	case err := <-errs:
		t.Fatalf("error on receiving prices: %v", err)
	case prices := <-data:
		assert.True(t, len(prices) > 0)
		price := prices[0]
		assert.True(t, price.Symbol != "")
		assert.True(t, price.TradeCause != "")
		assert.True(t, price.Timestamp > 0)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for prices")
	}
}
