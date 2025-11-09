package pacifica

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRESTClient_GetMarketInfo(t *testing.T) {
	client := NewRESTClient(MainnetAPIURL, nil)
	ctx := context.Background()

	marketInfo, err := client.GetMarketInfo(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, marketInfo)
	assert.NotEqual(t, len(marketInfo), 0)
}
