package pacifica

import (
	"encoding/json"
)

const (
	ChannelOrderBook = "book"
)

type wsCommand struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

type wsMessage struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type (
	OrderBook struct {
		Coin   string    `json:"s"`
		Levels [][]Level `json:"l"`
		Time   int64     `json:"t"`
	}

	Level struct {
		Quantity string `json:"a"` //Total amount in aggregation level.
		Price    string `json:"p"` //Price level.
		Orders   int    `json:"n"` // Number of orders in aggregation level.
	}
)
