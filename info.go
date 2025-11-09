package pacifica

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type marketInfoResponse struct {
	Success bool         `json:"success"`
	Data    []SymbolInfo `json:"data"`
	Error   interface{}  `json:"error"`
	Code    interface{}  `json:"code"`
}

type SymbolInfo struct {
	Symbol          string `json:"symbol"`
	TickSize        string `json:"tick_size"`
	MinTick         string `json:"min_tick"`
	MaxTick         string `json:"max_tick"`
	LotSize         string `json:"lot_size"`
	MaxLeverage     int    `json:"max_leverage"`
	IsolatedOnly    bool   `json:"isolated_only"`
	MinOrderSize    string `json:"min_order_size"`
	MaxOrderSize    string `json:"max_order_size"`
	FundingRate     string `json:"funding_rate"`
	NextFundingRate string `json:"next_funding_rate"`
}

func (c *RESTClient) GetMarketInfo(ctx context.Context) ([]SymbolInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/info", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("info: error creating request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("info: error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info: unexpected status code: %d", resp.StatusCode)
	}

	var marketInfoResp marketInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&marketInfoResp)
	if err != nil {
		return nil, fmt.Errorf("info: error decoding response: %w", err)
	}
	if !marketInfoResp.Success {
		return nil, fmt.Errorf("info: api error: %v", marketInfoResp.Error)
	}
	return marketInfoResp.Data, nil
}
