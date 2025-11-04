# Go Pacifica SDK

A Go SDK for interacting with the Pacifica Exchange API. This SDK provides both REST API and WebSocket functionality for trading, market data, and account management.

## Features

### REST API
- ✅ **Authentication & Signing**
  - Ed25519 signature generation
  - Automatic request signing with timestamp and expiry windows
  - Signature verification
  
- ✅ **Order Management**
  - Create limit orders with optional take profit and stop loss
  - Create market orders with slippage control
  - Cancel orders by order ID or client order ID
  
- ✅ **Request Building**
  - Automatic JSON key sorting for deterministic signatures
  - Compact JSON generation
  - Request validation

### WebSocket API
- ✅ **Real-time Market Data**
  - Order book subscriptions
  - Price updates (mark, mid, funding, oracle, etc.)
  - Trade stream subscriptions
  - Candle/OHLCV data subscriptions
  
- ✅ **Connection Management**
  - Automatic reconnection
  - Ping/pong keepalive
  - Subscription management

## Installation

```bash
go get github.com/KushnerykPavel/go-pacifica
```

## Dependencies

- `github.com/gorilla/websocket` - WebSocket client
- `github.com/mr-tron/base58` - Base58 encoding/decoding
- `github.com/sonirico/vago` - Utility functions
- `github.com/stretchr/testify` - Testing framework

## Quick Start

### 1. Authentication Setup

First, create an `Exchange` instance with your private key and account ID:

```go
package main

import (
    "fmt"
    "github.com/KushnerykPavel/go-pacifica"
)

func main() {
    // Your base58-encoded private key
    privateKey := "your_private_key_here"
    accountID := "your_account_id_here"
    
    // Create exchange instance
    exchange, err := pacifica.NewExchange(privateKey, accountID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Public Key: %s\n", exchange.GetPublicKey())
}
```

### 2. REST API Client

Create a REST client for making API calls:

```go
// Create REST client (uses mainnet by default)
client := pacifica.NewRESTClient("", exchange)

// Or use a custom base URL
client := pacifica.NewRESTClient("https://api.pacifica.fi/api/v1", exchange)
```

## Code Examples

### Create Limit Order

```go
// Build limit order request
params := pacifica.CreateLimitOrderRequest{
    Symbol:     "BTC",
    Price:      "50000",
    Amount:     "0.1",
    Side:       pacifica.SideBid,
    TIF:        pacifica.TIFGTC,
    ReduceOnly: false,
    TakeProfit: &pacifica.Target{
        StopPrice:     "55000",
        LimitPrice:    "54950",
        ClientOrderID: "tp-client-id-123",
    },
    StopLoss: &pacifica.Target{
        StopPrice:     "48000",
        LimitPrice:    "47950",
        ClientOrderID: "sl-client-id-456",
    },
}

// Optional parameters
opts := &pacifica.CreateLimitOrderOptions{
    AgentWallet:  nil, // Optional agent wallet
    ExpiryWindow: 30000, // 30 seconds
}

// Create the order
response, err := client.CreateLimitOrder(params, opts)
if err != nil {
    fmt.Printf("Error creating order: %v\n", err)
    return
}

fmt.Printf("Order created with ID: %d\n", response.OrderID)
```

### Create Market Order

```go
// Build market order request
params := pacifica.CreateMarketOrderRequest{
    Symbol:          "BTC",
    Amount:          "0.1",
    Side:            pacifica.SideBid,
    SlippagePercent: "0.5", // 0.5% max slippage
    ReduceOnly:      false,
    ClientOrderID:   "market-order-123",
}

// Create the market order
response, err := client.CreateMarketOrder(params, nil)
if err != nil {
    fmt.Printf("Error creating market order: %v\n", err)
    return
}

fmt.Printf("Market order created with ID: %d\n", response.OrderID)
```

### Cancel Order

Cancel by order ID:

```go
params := pacifica.CancelOrderRequest{
    Symbol:  "BTC",
    OrderID: intPtr(12345), // Order ID from exchange
}

response, err := client.CancelOrder(params, nil)
if err != nil {
    fmt.Printf("Error canceling order: %v\n", err)
    return
}

if response.Success {
    fmt.Println("Order cancelled successfully")
}
```

Cancel by client order ID:

```go
params := pacifica.CancelOrderRequest{
    Symbol:        "BTC",
    ClientOrderID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
}

response, err := client.CancelOrder(params, nil)
if err != nil {
    fmt.Printf("Error canceling order: %v\n", err)
    return
}
```

### WebSocket Subscriptions

#### Order Book

```go
package main

import (
    "context"
    "fmt"
    "github.com/KushnerykPavel/go-pacifica"
)

func main() {
    // Create WebSocket client
    wsClient := pacifica.NewWebsocketClient("")
    
    // Connect to WebSocket
    ctx := context.Background()
    if err := wsClient.Connect(ctx); err != nil {
        panic(err)
    }
    defer wsClient.Close()
    
    // Subscribe to order book
    sub, err := wsClient.OrderBook("BTC", func(book pacifica.OrderBook) {
        fmt.Printf("Order Book Update for %s:\n", book.Coin)
        fmt.Printf("Time: %d\n", book.Time)
        fmt.Printf("Bids: %+v\n", book.Levels[0]) // Bids
        fmt.Printf("Asks: %+v\n", book.Levels[1]) // Asks
    })
    if err != nil {
        panic(err)
    }
    
    // Keep subscription alive
    select {}
    
    // Unsubscribe when done
    sub.Close()
}
```

#### Price Updates

```go
// Subscribe to price updates
sub, err := wsClient.Prices("BTC", func(prices pacifica.Prices) {
    for _, price := range prices {
        fmt.Printf("Symbol: %s\n", price.Symbol)
        fmt.Printf("Mark Price: %s\n", price.Mark)
        fmt.Printf("Mid Price: %s\n", price.Mid)
        fmt.Printf("Funding Rate: %s\n", price.Funding)
        fmt.Printf("Oracle Price: %s\n", price.Oracle)
        fmt.Printf("24h Volume: %s\n", price.Volume24H)
    }
})
if err != nil {
    panic(err)
}
defer sub.Close()
```

#### Trade Stream

```go
// Subscribe to trades
sub, err := wsClient.Trades("BTC", func(trades pacifica.Trades) {
    for _, trade := range trades {
        fmt.Printf("Trade: %s @ %s, Amount: %s, Side: %s\n",
            trade.Price, trade.Timestamp, trade.Amount, trade.TradeSide)
    }
})
if err != nil {
    panic(err)
}
defer sub.Close()
```

#### Candle/OHLCV Data

```go
// Subscribe to candle data
sub, err := wsClient.Candle("BTC", "1m", func(candle pacifica.Candle) {
    fmt.Printf("Candle for %s:\n", candle.Symbol)
    fmt.Printf("Interval: %s\n", candle.Interval)
    fmt.Printf("Open: %s, High: %s, Low: %s, Close: %s\n",
        candle.Open, candle.High, candle.Low, candle.Close)
    fmt.Printf("Volume: %s, Trades: %d\n", candle.Volume, candle.NumberTrades)
})
if err != nil {
    panic(err)
}
defer sub.Close()
```

### Advanced: Building Signed Requests Manually

You can also build signed requests manually without using the REST client:

```go
// Build a signed request for a limit order
params := pacifica.CreateLimitOrderRequest{
    Symbol:     "BTC",
    Price:      "50000",
    Amount:     "0.1",
    Side:       pacifica.SideBid,
    TIF:        pacifica.TIFGTC,
    ReduceOnly: false,
}

request, err := exchange.BuildCreateLimitOrderRequest(params, nil)
if err != nil {
    panic(err)
}

// Use the request map to make your own HTTP call
fmt.Printf("Signed request: %+v\n", request)
```

### Error Handling

```go
response, err := client.CreateLimitOrder(params, nil)
if err != nil {
    // Check if it's an API error
    if strings.Contains(err.Error(), "API error") {
        // Handle API-specific errors
        fmt.Printf("API Error: %v\n", err)
    } else {
        // Handle other errors (network, validation, etc.)
        fmt.Printf("Error: %v\n", err)
    }
    return
}

// Success
fmt.Printf("Order ID: %d\n", response.OrderID)
```

### Helper Functions

```go
// Helper to create int pointer
func intPtr(i int64) *int64 {
    return &i
}

// Helper to create string pointer
func stringPtr(s string) *string {
    return &s
}
```

## Testing

Run all tests:

```bash
go test ./...
```

Run specific test suites:

```bash
# Test REST API
go test -v -run TestBuildCreateLimitOrderRequest
go test -v -run TestBuildCreateMarketOrderRequest
go test -v -run TestBuildCancelOrderRequest

# Test WebSocket
go test -v -run TestOrderBook
go test -v -run TestPrices
go test -v -run TestTrades
go test -v -run TestCandle

# Test authentication
go test -v -run TestCreateSignature
```

## API Documentation

For detailed API documentation, refer to:
- [Pacifica API Documentation](https://docs.pacifica.fi/api-documentation)
- [REST API Reference](https://docs.pacifica.fi/api-documentation/api/rest-api)
- [WebSocket API Reference](https://docs.pacifica.fi/api-documentation/api/websocket)
- [Signing Implementation](https://docs.pacifica.fi/api-documentation/api/signing/implementation)

## Constants

### Order Sides
- `pacifica.SideBid` - Buy side
- `pacifica.SideAsk` - Sell side

### Time in Force
- `pacifica.TIFGTC` - Good Till Cancel
- `pacifica.TIFIOC` - Immediate Or Cancel
- `pacifica.TIFALO` - Allow Override

### API URLs
- `pacifica.MainnetAPIURL` - Mainnet REST API URL
- `pacifica.MainnetWSURL` - Mainnet WebSocket URL

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [Pacifica API Documentation](https://docs.pacifica.fi/api-documentation)

