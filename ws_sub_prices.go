package pacifica

import (
	"fmt"
)

func (w *WebsocketClient) Prices(
	callback func(Prices, error),
) (*Subscription, error) {
	remotePayload := remotePricesSubscriptionPayload{
		Source: ChannelPrices,
	}

	return w.subscribe(remotePayload, func(msg any) {
		prices, ok := msg.(Prices)
		if !ok {
			callback(Prices{}, fmt.Errorf("invalid message type"))
			return
		}
		callback(prices, nil)
	})
}
