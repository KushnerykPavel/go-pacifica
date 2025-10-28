package pacifica

import (
	"errors"
	"fmt"
	"slices"
)

var validIntervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "8h", "12h", "1d"}

type CandleSubscriptionParams struct {
	Symbol   string
	Interval string
}

func (w *WebsocketClient) Candle(
	params CandleSubscriptionParams,
	callback func(Candle, error),
) (*Subscription, error) {
	if !slices.Contains(validIntervals, params.Interval) {
		return nil, fmt.Errorf("invalid interval: %s", params.Interval)
	}

	remotePayload := remoteCandleSubscriptionPayload{
		Source:   ChannelCandle,
		Symbol:   params.Symbol,
		Interval: params.Interval,
	}
	return w.subscribe(remotePayload, func(msg any) {
		candles, ok := msg.(Candle)
		if !ok {
			callback(Candle{}, errors.New("invalid message type"))
			return
		}
		callback(candles, nil)
	})
}
