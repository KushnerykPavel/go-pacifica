package pacifica

import (
	"encoding/json"
)

const (
	ChannelPrices      = "prices"
	ChannelPong        = "pong"
	ChannelOrderBook   = "book"
	ChannelTrades      = "trades"
	ChannelCandle      = "candle"
	ChannelSubResponse = "subscribe"
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

	Price struct {
		Funding        string `json:"funding"`
		Mark           string `json:"mark"`
		Mid            string `json:"mid"`
		NextFunding    string `json:"next_funding"`
		OpenInterest   string `json:"open_interest"`
		Oracle         string `json:"oracle"`
		Symbol         string `json:"symbol"`
		Timestamp      int64  `json:"timestamp"`
		Volume24H      string `json:"volume_24h"`
		YesterdayPrice string `json:"yesterday_price"`
	}

	Prices []Price

	Trade struct {
		HistoryID      int    `json:"h"`
		Amount         string `json:"a"`
		TradeSide      string `json:"d"`
		Price          string `json:"p"`
		Symbol         string `json:"s"`
		Timestamp      int64  `json:"t"`
		TradeCause     string `json:"tc"`
		AccountAddress string `json:"u"`
	}

	Trades []Trade

	Candle struct {
		StartTime    int64  `json:"t"`
		EndTime      int64  `json:"T"`
		Symbol       string `json:"s"`
		Interval     string `json:"i"`
		Open         string `json:"o"`
		Close        string `json:"c"`
		High         string `json:"h"`
		Low          string `json:"l"`
		Volume       string `json:"v"`
		NumberTrades int    `json:"n"`
	}
)
