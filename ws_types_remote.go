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
