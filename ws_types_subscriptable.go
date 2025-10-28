package pacifica

type subscriptable interface {
	Key() string
}

func (c OrderBook) Key() string {
	return keyOrderBook(c.Coin)
}
