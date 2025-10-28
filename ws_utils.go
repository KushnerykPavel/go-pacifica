package pacifica

import (
	"strings"
)

func key(args ...string) string {
	return strings.Join(args, ":")
}

func keyOrderBook(coin string) string {
	return key(ChannelOrderBook, coin)
}

func keyTrades(coin string) string {
	return key(ChannelTrades, coin)
}

func keyPrices() string {
	return key(ChannelPrices)
}

func keyCandle(coin, interval string) string {
	return key(ChannelCandle, coin, interval)
}
