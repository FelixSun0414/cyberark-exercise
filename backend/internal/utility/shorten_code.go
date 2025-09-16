package utility

// Base62 table(0-9A-Za-z)
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// pre-defined 62^k constants(k ∈ {5..8})
const (
	Pow62_5 uint64 = 916132832
	Pow62_6 uint64 = 56800235584
	Pow62_7 uint64 = 3521614606208
	Pow62_8 uint64 = 218340105584896
)

// Affine constants
// Note: DO NOT change these values once the service is released.
// With A=84485 and B=1e9+7, A*x+B won't overflow uint64 when x < 62^8.
const (
	Affine_A uint64 = 84485      // odd and not divisible by 31. it coprime with 62^k
	Affine_B uint64 = 1000000007 // just a random small constant
)

// EncodeNumberToShortenCode encodes an unsigned integer into a Base62 string.
// The output length k is chosen automatically by x (< 62^k), with a minimum of 5 and a maximum of 8.
// IMPORTANT: callers must ensure x < 62^8 to avoid wrap-around collisions.
func EncodeNumberToShortenCode(x uint64) (string, int) {
	k := getKForValue(x)
	M := pow62(k)

	// Affine permutation: y = (A*x + B) mod M
	y := (Affine_A*x + Affine_B) % M

	// Encode and rotate
	code := encodeFixedBase62(y, k)
	return rotateRightBy1(code), k
}

// encodeFixedBase62 encodes n as a fixed-length Base62 string of length k.
// Precondition: n < 62^k and 5 <= k <= 8.
func encodeFixedBase62(n uint64, k int) string {
	b := make([]byte, k)
	for i := k - 1; i >= 0; i-- {
		b[i] = alphabet[n%62]
		n /= 62
	}
	return string(b)
}

// rotateRightBy1 moves the last character to the front to make the first char looks more "dynamic"
func rotateRightBy1(s string) string {
	if len(s) < 2 {
		return s
	}
	b := []byte(s)
	last := b[len(b)-1]
	copy(b[1:], b[:len(b)-1])
	b[0] = last
	return string(b)
}

// pow62 gets the maximum number based on the length of the expected code
func pow62(k int) uint64 {
	switch k {
	case 5:
		return Pow62_5
	case 6:
		return Pow62_6
	case 7:
		return Pow62_7
	case 8:
		return Pow62_8
	default:
		return Pow62_8
	}
}

// getKForValue picks the minimal k ∈ {5..8} such that x < 62^k.
// If x ≥ 62^8, it clamps to k=8 (caller must ensure x < 62^8 to avoid wrap-around).
func getKForValue(x uint64) int {
	switch {
	case x < Pow62_5:
		return 5
	case x < Pow62_6:
		return 6
	case x < Pow62_7:
		return 7
	case x < Pow62_8:
		return 8
	default:
		// 8 will be the max length for the code, it's actually impossible to use out Pow62_8
		// For constant POST QPS = 5k, Pow62_8 is good for 1384 years
		return 8
	}
}
