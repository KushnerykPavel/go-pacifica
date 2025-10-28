package pacifica_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KushnerykPavel/go-pacifica"
)

func TestWebsocketClient_Candle(t *testing.T) {
	client := pacifica.NewWebsocketClient(pacifica.MainnetWSURL)

	err := client.Connect(context.Background())
	assert.NoError(t, err)
	data := make(chan pacifica.Candle)
	errs := make(chan error)
	_, _ = client.Candle(pacifica.CandleSubscriptionParams{
		Symbol:   "SOL",
		Interval: "1m",
	}, func(candle pacifica.Candle, err error) {
		if err != nil {
			errs <- err
			return
		}
		data <- candle
	})

	select {
	case err := <-errs:
		t.Fatalf("error on receiving candle: %v", err)
	case candle := <-data:
		assert.True(t, candle.Symbol != "")
		assert.True(t, candle.Interval != "")
		assert.True(t, candle.StartTime > 0)
		assert.True(t, candle.EndTime > 0)
	case <-time.After(62 * time.Second):
		t.Fatal("timeout waiting for candle")
	}
}
