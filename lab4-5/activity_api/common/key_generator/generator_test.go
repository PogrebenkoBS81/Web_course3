package key_generator

import (
	"encoding/pem"
	"testing"
)

// TestService_GenerateKey - tests GenerateKey
func TestService_GenerateKey(t *testing.T) {
	t.Run("GenerateKey_test", func(t *testing.T) {
		private, public, err := GenerateKey()

		if err != nil {
			t.Fatal(err)
		}

		_, rest := pem.Decode([]byte(private))
		// If there is an extra data after -----END PRIVATE KEY----- - throw an error.
		if len(rest) > 0 {
			t.Fatalf("unexpected extra block length: %d", len(rest))
		}

		_, rest = pem.Decode([]byte(public))
		// If there is an extra data after -----END PUBLIC KEY----- - throw an error.
		if len(rest) > 0 {
			t.Fatalf("unexpected extra block length: %d", len(rest))
		}
	})
}
