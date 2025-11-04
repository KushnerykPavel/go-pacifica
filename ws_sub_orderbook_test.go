package pacifica_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KushnerykPavel/go-pacifica"
)

type debugLogger struct{}

func (d debugLogger) Infof(format string, args ...any) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (d debugLogger) Errorf(format string, args ...any) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func TestWebsocketClient_OrderBook(t *testing.T) {
	client := pacifica.NewWebsocketClient(pacifica.MainnetWSURL)

	err := client.Connect(context.Background())
	assert.NoError(t, err)

	data := make(chan pacifica.OrderBook)
	errs := make(chan error)

	_, _ = client.OrderBook(pacifica.OrderBookSubscriptionParams{Symbol: "SOL", AggLevel: 1}, func(book pacifica.OrderBook, err error) {
		if err != nil {
			errs <- err
			return
		}
		data <- book
	})

	select {
	case err := <-errs:
		t.Fatalf("error on receiving prices: %v", err)
	case ob := <-data:
		assert.True(t, ob.Coin != "")
		assert.True(t, len(ob.Levels[0]) != 0)
		assert.True(t, len(ob.Levels[1]) != 0)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for prices")
	}
}
