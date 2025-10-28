package pacifica

import (
	"fmt"
)

type TradesSubscriptionParams struct {
	Symbol string
}

func (w *WebsocketClient) Trades(
	params TradesSubscriptionParams,
	callback func(Trades, error),
) (*Subscription, error) {
	remotePayload := remoteTradesSubscriptionPayload{
		Source: ChannelTrades,
		Symbol: params.Symbol,
	}
	return w.subscribe(remotePayload, func(msg any) {
		trades, ok := msg.(Trades)
		if !ok {
			callback(Trades{}, fmt.Errorf("invalid message type"))
			return
		}
		callback(trades, nil)
	})
}
