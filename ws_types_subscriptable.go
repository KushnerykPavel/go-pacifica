package pacifica

type subscriptable interface {
	Key() string
}

func (c OrderBook) Key() string {
	return keyOrderBook(c.Coin)
}

func (c Prices) Key() string {
	return keyPrices()
}

func (c Trades) Key() string {
	if len(c) == 0 {
		return ""
	}
	return keyTrades(c[0].Symbol)
}

func (c Candle) Key() string {
	return keyCandle(c.Symbol, c.Interval)
}
