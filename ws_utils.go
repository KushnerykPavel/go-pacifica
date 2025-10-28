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
