package pacifica

type remoteOrderBookSubscriptionPayload struct {
	Source   string `json:"source"`
	Symbol   string `json:"symbol"`
	AggLevel int    `json:"agg_level,omitempty"`
}

func (p remoteOrderBookSubscriptionPayload) Channel() string {
	return p.Source
}

func (p remoteOrderBookSubscriptionPayload) Key() string {
	return keyOrderBook(p.Symbol)
}

type remotePricesSubscriptionPayload struct {
	Source string `json:"source"`
}

func (p remotePricesSubscriptionPayload) Channel() string {
	return p.Source
}

func (p remotePricesSubscriptionPayload) Key() string {
	return keyPrices()
}

type remoteTradesSubscriptionPayload struct {
	Source string `json:"source"`
	Symbol string `json:"symbol"`
}

func (p remoteTradesSubscriptionPayload) Channel() string {
	return p.Source
}

func (p remoteTradesSubscriptionPayload) Key() string {
	return keyTrades(p.Symbol)
}

type remoteCandleSubscriptionPayload struct {
	Source   string `json:"source"`
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
}

func (p remoteCandleSubscriptionPayload) Channel() string {
	return p.Source
}

func (p remoteCandleSubscriptionPayload) Key() string {
	return keyCandle(p.Symbol, p.Interval)
}
