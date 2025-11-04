package pacifica

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mr-tron/base58"
)

// SignatureHeader represents the header structure for Pacifica API signing
type SignatureHeader struct {
	Timestamp    int64  `json:"timestamp"`
	ExpiryWindow int64  `json:"expiry_window"`
	Type         string `json:"type"`
}

// SignedRequest represents the final request structure with authentication
type SignedRequest struct {
	Account      string      `json:"account"`
	AgentWallet  *string     `json:"agent_wallet"`
	Signature    string      `json:"signature"`
	Timestamp    int64       `json:"timestamp"`
	ExpiryWindow int64       `json:"expiry_window"`
	Data         interface{} `json:"-"` // This will be flattened into the request
}

// Exchange handles Pacifica API signature generation
type Exchange struct {
	accountID  string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewExchange creates a new signer instance from a base58 encoded private key
func NewExchange(privateKeyBase58 string, accountID string) (*Exchange, error) {
	// Decode base58 private key
	privateKeyBytes, err := base58.Decode(privateKeyBase58)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Convert to ed25519 private key
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Exchange{
		accountID:  accountID,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// GetPublicKey returns the base58 encoded public key
func (s *Exchange) GetPublicKey() string {
	return base58.Encode(s.publicKey)
}

// sortJSONKeys recursively sorts all keys in a JSON structure
func sortJSONKeys(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		// Create a new map with sorted keys
		sortedMap := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sortedMap[k] = sortJSONKeys(v[k])
		}
		return sortedMap
	case []interface{}:
		// Sort array elements recursively
		sortedArray := make([]interface{}, len(v))
		for i, item := range v {
			sortedArray[i] = sortJSONKeys(item)
		}
		return sortedArray
	default:
		return v
	}
}

// createCompactJSON creates a compact JSON string with no whitespace
func createCompactJSON(data interface{}) (string, error) {
	// Sort the JSON keys recursively
	sortedData := sortJSONKeys(data)

	// Marshal to compact JSON
	jsonBytes, err := json.Marshal(sortedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// signMessage signs a message using the private key
func (s *Exchange) signMessage(message string) (string, error) {
	// Convert message to bytes
	messageBytes := []byte(message)

	// Sign the message
	signature := ed25519.Sign(s.privateKey, messageBytes)

	// Convert signature to base58
	signatureBase58 := base58.Encode(signature)

	return signatureBase58, nil
}

// CreateSignature creates a signature for the given operation data
func (s *Exchange) CreateSignature(operationType string, operationData interface{}, expiryWindow int64) (*SignatureHeader, string, error) {
	// Get current timestamp in milliseconds
	timestamp := time.Now().UnixMilli()

	// Use default expiry window if not provided
	if expiryWindow == 0 {
		expiryWindow = 30000 // 30 seconds default
	}

	// Create signature header
	header := &SignatureHeader{
		Timestamp:    timestamp,
		ExpiryWindow: expiryWindow,
		Type:         operationType,
	}

	// Combine header and payload
	dataToSign := map[string]interface{}{
		"timestamp":     header.Timestamp,
		"expiry_window": header.ExpiryWindow,
		"type":          header.Type,
		"data":          operationData,
	}

	// Create compact JSON
	compactJSON, err := createCompactJSON(dataToSign)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create compact JSON: %w", err)
	}

	// Sign the message
	signature, err := s.signMessage(compactJSON)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign message: %w", err)
	}

	return header, signature, nil
}

// BuildSignedRequest builds the final request with authentication headers
func (s *Exchange) BuildSignedRequest(operationType string, operationData interface{}, expiryWindow int64) (map[string]interface{}, error) {
	// Create signature
	header, signature, err := s.CreateSignature(operationType, operationData, expiryWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	// Convert operation data to map for flattening
	var dataMap map[string]interface{}
	if operationData != nil {
		jsonBytes, err := json.Marshal(operationData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal operation data: %w", err)
		}

		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal operation data: %w", err)
		}
	} else {
		dataMap = make(map[string]interface{})
	}

	// Build final request
	request := map[string]interface{}{
		"account":       s.accountID,
		"agent_wallet":  s.GetPublicKey(),
		"signature":     signature,
		"timestamp":     header.Timestamp,
		"expiry_window": header.ExpiryWindow,
	}

	// Flatten operation data into the request
	for k, v := range dataMap {
		request[k] = v
	}

	return request, nil
}

// VerifySignature verifies a signature against a message
func (s *Exchange) VerifySignature(message, signature string) bool {
	// Decode base58 signature
	signatureBytes, err := base58.Decode(signature)
	if err != nil {
		return false
	}

	// Verify the signature
	return ed25519.Verify(s.publicKey, []byte(message), signatureBytes)
}
