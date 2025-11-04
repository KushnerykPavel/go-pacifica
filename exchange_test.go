package pacifica

import (
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountID = "test_account_id"

// generateTestExchange creates an Exchange with a valid ed25519 private key for testing
func generateTestExchange(t *testing.T) *Exchange {
	_, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	privateKeyBase58 := base58.Encode(privateKey)
	signer, err := NewExchange(privateKeyBase58, testAccountID)
	require.NoError(t, err)
	return signer
}

func TestNewSigner(t *testing.T) {
	// Test with valid private key
	signer := generateTestExchange(t)
	assert.NotNil(t, signer)
	assert.NotEmpty(t, signer.GetPublicKey())

	// Test with invalid private key
	_, err := NewExchange("invalid_key", testAccountID)
	assert.Error(t, err)
}

func TestSortJSONKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"c": "value3",
				"a": "value1",
				"b": "value2",
			},
			expected: map[string]interface{}{
				"a": "value1",
				"b": "value2",
				"c": "value3",
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"z": map[string]interface{}{
					"c": "nested3",
					"a": "nested1",
					"b": "nested2",
				},
				"a": "value1",
			},
			expected: map[string]interface{}{
				"a": "value1",
				"z": map[string]interface{}{
					"a": "nested1",
					"b": "nested2",
					"c": "nested3",
				},
			},
		},
		{
			name: "array with maps",
			input: []interface{}{
				map[string]interface{}{
					"c": "value3",
					"a": "value1",
				},
				map[string]interface{}{
					"b": "value2",
					"a": "value1",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"a": "value1",
					"c": "value3",
				},
				map[string]interface{}{
					"a": "value1",
					"b": "value2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortJSONKeys(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateCompactJSON(t *testing.T) {
	data := map[string]interface{}{
		"timestamp":     1748970123456,
		"expiry_window": 5000,
		"type":          "create_order",
		"data": map[string]interface{}{
			"symbol":          "BTC",
			"price":           "100000",
			"amount":          "0.1",
			"side":            "bid",
			"tif":             "GTC",
			"reduce_only":     false,
			"client_order_id": "12345678-1234-1234-1234-123456789abc",
		},
	}

	compactJSON, err := createCompactJSON(data)
	require.NoError(t, err)

	// Verify it's compact (no spaces)
	assert.NotContains(t, compactJSON, " ")
	assert.NotContains(t, compactJSON, "\n")
	assert.NotContains(t, compactJSON, "\t")

	// Verify it can be unmarshaled back
	var result map[string]interface{}
	err = json.Unmarshal([]byte(compactJSON), &result)
	require.NoError(t, err)

	// Verify the structure is preserved
	assert.Equal(t, "create_order", result["type"])
}

func TestCreateSignature(t *testing.T) {
	signer := generateTestExchange(t)

	operationData := map[string]interface{}{
		"symbol":          "BTC",
		"price":           "100000",
		"amount":          "0.1",
		"side":            "bid",
		"tif":             "GTC",
		"reduce_only":     false,
		"client_order_id": "12345678-1234-1234-1234-123456789abc",
	}

	header, signature, err := signer.CreateSignature("create_order", operationData, 5000)
	require.NoError(t, err)

	// Verify header
	assert.Equal(t, "create_order", header.Type)
	assert.Equal(t, int64(5000), header.ExpiryWindow)
	assert.True(t, header.Timestamp > 0)

	// Verify signature is not empty
	assert.NotEmpty(t, signature)

	// Verify signature can be verified
	// Recreate the data that was signed
	dataToSign := map[string]interface{}{
		"timestamp":     header.Timestamp,
		"expiry_window": header.ExpiryWindow,
		"type":          header.Type,
		"data":          operationData,
	}
	compactJSON, err := createCompactJSON(dataToSign)
	require.NoError(t, err)

	verified := signer.VerifySignature(compactJSON, signature)
	assert.True(t, verified)
}

func TestBuildSignedRequest(t *testing.T) {
	signer := generateTestExchange(t)

	operationData := map[string]interface{}{
		"symbol":          "BTC",
		"price":           "100000",
		"amount":          "0.1",
		"side":            "bid",
		"tif":             "GTC",
		"reduce_only":     false,
		"client_order_id": "12345678-1234-1234-1234-123456789abc",
	}

	request, err := signer.BuildSignedRequest("create_order", operationData, 5000)
	require.NoError(t, err)

	// Verify required fields are present
	assert.Contains(t, request, "account")
	assert.Contains(t, request, "agent_wallet")
	assert.Contains(t, request, "signature")
	assert.Contains(t, request, "timestamp")
	assert.Contains(t, request, "expiry_window")

	// Verify operation data is flattened
	assert.Contains(t, request, "symbol")
	assert.Contains(t, request, "price")
	assert.Contains(t, request, "amount")
	assert.Contains(t, request, "side")
	assert.Contains(t, request, "tif")
	assert.Contains(t, request, "reduce_only")
	assert.Contains(t, request, "client_order_id")

	// Verify values
	assert.Equal(t, "BTC", request["symbol"])
	assert.Equal(t, "100000", request["price"])
	assert.Equal(t, "0.1", request["amount"])
	assert.Equal(t, "bid", request["side"])
	assert.Equal(t, "GTC", request["tif"])
	assert.Equal(t, false, request["reduce_only"])
	assert.Equal(t, "12345678-1234-1234-1234-123456789abc", request["client_order_id"])
	assert.Equal(t, testAccountID, request["account"])
	assert.NotEmpty(t, request["agent_wallet"])
	assert.IsType(t, "", request["agent_wallet"])

	// Verify signature is valid
	compactJSON, err := createCompactJSON(map[string]interface{}{
		"timestamp":     request["timestamp"],
		"expiry_window": request["expiry_window"],
		"type":          "create_order",
		"data":          operationData,
	})
	require.NoError(t, err)

	verified := signer.VerifySignature(compactJSON, request["signature"].(string))
	assert.True(t, verified)
}

func TestDefaultExpiryWindow(t *testing.T) {
	signer := generateTestExchange(t)

	operationData := map[string]interface{}{
		"symbol": "BTC",
		"price":  "100000",
	}

	header, _, err := signer.CreateSignature("create_order", operationData, 0)
	require.NoError(t, err)

	// Should use default 30 seconds
	assert.Equal(t, int64(30000), header.ExpiryWindow)
}

func TestVerifySignature(t *testing.T) {
	signer := generateTestExchange(t)

	message := "test message"
	signature, err := signer.signMessage(message)
	require.NoError(t, err)

	// Verify correct signature
	assert.True(t, signer.VerifySignature(message, signature))

	// Verify wrong message
	assert.False(t, signer.VerifySignature("wrong message", signature))

	// Verify wrong signature
	assert.False(t, signer.VerifySignature(message, "wrong_signature"))
}

func TestSignatureConsistency(t *testing.T) {
	signer := generateTestExchange(t)

	operationData := map[string]interface{}{
		"symbol": "BTC",
		"price":  "100000",
		"amount": "0.1",
	}

	// Create multiple signatures with the same data
	// They should be different due to different timestamps
	header1, sig1, err := signer.CreateSignature("create_order", operationData, 5000)
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond) // Ensure different timestamp

	header2, sig2, err := signer.CreateSignature("create_order", operationData, 5000)
	require.NoError(t, err)

	// Signatures should be different due to different timestamps
	assert.NotEqual(t, sig1, sig2)
	assert.NotEqual(t, header1.Timestamp, header2.Timestamp)

	// But both should be valid
	compactJSON1, err := createCompactJSON(map[string]interface{}{
		"timestamp":     header1.Timestamp,
		"expiry_window": header1.ExpiryWindow,
		"type":          "create_order",
		"data":          operationData,
	})
	require.NoError(t, err)

	compactJSON2, err := createCompactJSON(map[string]interface{}{
		"timestamp":     header2.Timestamp,
		"expiry_window": header2.ExpiryWindow,
		"type":          "create_order",
		"data":          operationData,
	})
	require.NoError(t, err)

	assert.True(t, signer.VerifySignature(compactJSON1, sig1))
	assert.True(t, signer.VerifySignature(compactJSON2, sig2))
}

func TestComplexOperationData(t *testing.T) {
	signer := generateTestExchange(t)

	// Test with complex nested data
	operationData := map[string]interface{}{
		"symbol":          "ETH",
		"price":           "2000",
		"amount":          "1.5",
		"side":            "ask",
		"tif":             "IOC",
		"reduce_only":     true,
		"client_order_id": "complex-test-123",
		"metadata": map[string]interface{}{
			"source":  "api",
			"version": "1.0",
			"tags":    []interface{}{"test", "complex"},
		},
	}

	request, err := signer.BuildSignedRequest("create_order", operationData, 10000)
	require.NoError(t, err)

	// Verify all fields are present
	assert.Equal(t, "ETH", request["symbol"])
	assert.Equal(t, "2000", request["price"])
	assert.Equal(t, "1.5", request["amount"])
	assert.Equal(t, "ask", request["side"])
	assert.Equal(t, "IOC", request["tif"])
	assert.Equal(t, true, request["reduce_only"])
	assert.Equal(t, "complex-test-123", request["client_order_id"])

	// Verify metadata is flattened
	assert.Contains(t, request, "metadata")
	metadata, ok := request["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "api", metadata["source"])
	assert.Equal(t, "1.0", metadata["version"])
	assert.Contains(t, metadata, "tags")
}
