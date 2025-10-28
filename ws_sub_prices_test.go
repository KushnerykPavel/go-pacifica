package pacifica_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KushnerykPavel/go-pacifica"
)

func TestWebsocketClient_Prices(t *testing.T) {
	client := pacifica.NewWebsocketClient(pacifica.MainnetWSURL)

	err := client.Connect(context.Background())
	assert.NoError(t, err)
	data := make(chan pacifica.Prices)
	errs := make(chan error)
	_, _ = client.Prices(func(prices pacifica.Prices, err error) {
		if err != nil {
			errs <- err
			return
		}
		data <- prices
	})

	select {
	case err := <-errs:
		t.Fatalf("error on receiving prices: %v", err)
	case prices := <-data:
		assert.True(t, len(prices) > 0)
		price := prices[0]
		assert.True(t, price.Symbol != "")
		assert.True(t, price.Mark != "")
		assert.True(t, price.Timestamp > 0)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for prices")
	}
}
