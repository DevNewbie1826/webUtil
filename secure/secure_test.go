package secure

import (
	"strings"
	"testing"
)

// BenchmarkCryptoRandNonce benchmarks the cryptoRandNonce function to measure performance and memory allocation.
func BenchmarkCryptoRandNonce(b *testing.B) {
	var buf strings.Builder
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cryptoRandNonce(&buf)
	}
}

// TestCryptoRandNonceBasic tests the basic functionality of cryptoRandNonce.
func TestCryptoRandNonceBasic(t *testing.T) {
	var buf strings.Builder
	cryptoRandNonce(&buf)
	nonce := buf.String()
	if len(nonce) == 0 {
		t.Errorf("expected non-empty nonce, got empty")
	}
	// Additional checks can be added, e.g., base64 validity
}

// To run memory profiling:
// 1. go test -bench=BenchmarkCryptoRandNonce -memprofile=mem.out
// 2. go tool pprof mem.out
// 3. In pprof, use commands like 'top' to see allocation hotspots, 'list cryptoRandNonce' for source-level details.
