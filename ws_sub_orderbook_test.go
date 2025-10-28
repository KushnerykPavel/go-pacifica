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

func TestWebsocketClient_L2Book(t *testing.T) {
	client := pacifica.NewWebsocketClient(pacifica.MainnetWSURL)

	err := client.Connect(context.Background())
	assert.NoError(t, err)

	_, _ = client.OrderBook(pacifica.OrderBookSubscriptionParams{Coin: "SOL", AggLevel: 1}, func(book pacifica.OrderBook, err error) {
		fmt.Println(book)
	})

	time.Sleep(time.Minute)
}
