package pacifica

import (
	"fmt"
)

type OrderBookSubscriptionParams struct {
	Coin     string
	AggLevel int
}

func (w *WebsocketClient) OrderBook(
	params OrderBookSubscriptionParams,
	callback func(OrderBook, error),
) (*Subscription, error) {
	remotePayload := remoteOrderBookSubscriptionPayload{
		Source:   ChannelOrderBook,
		Symbol:   params.Coin,
		AggLevel: params.AggLevel,
	}
	return w.subscribe(remotePayload, func(msg any) {
		orderbook, ok := msg.(OrderBook)
		if !ok {
			callback(OrderBook{}, fmt.Errorf("invalid message type"))
			return
		}
		callback(orderbook, nil)
	})
}
