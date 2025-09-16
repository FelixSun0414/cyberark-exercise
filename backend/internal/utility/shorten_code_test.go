package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlphabet(t *testing.T) {
	require.Equal(t, 62, len(alphabet), "alphabet length must be 62")

	seen := make(map[byte]bool, 62)
	for i := 0; i < len(alphabet); i++ {
		ch := alphabet[i]
		assert.Falsef(t, seen[ch], "duplicate char in alphabet: %q", ch)
		seen[ch] = true
	}
}

func TestEncodeNumberToShortenCode_DeterminismAndFormula(t *testing.T) {
	samples := []uint64{
		0, 1, 61, 62, 12345,
		Pow62_5 - 2, Pow62_5 - 1, Pow62_5,
		Pow62_6 - 1, Pow62_6,
		Pow62_7 - 1, Pow62_7,
		Pow62_8 - 1,
	}
	allowed := make(map[byte]bool, 62)
	for i := 0; i < len(alphabet); i++ {
		allowed[alphabet[i]] = true
	}

	for _, x := range samples {
		code, k := EncodeNumberToShortenCode(x)
		require.Equalf(t, getKForValue(x), k, "x=%d", x)

		// verify the length of the code is expected
		assert.Lenf(t, code, k, "x=%d", x)

		// verify the code only contains chars in alphabet
		for i := 0; i < k; i++ {
			_, ok := allowed[code[i]]
			assert.Truef(t, ok, "code has invalid char: %q (x=%d)", code[i], x)
		}

		// ensure the encoding follows the expected formula：rotateRightBy1(encodeFixedBase62((A*x+B)%M, k))
		M := pow62(k)
		y := (Affine_A*x + Affine_B) % M
		want := rotateRightBy1(encodeFixedBase62(y, k))
		assert.Equalf(t, want, code, "x=%d", x)
	}
}

func TestEncodeNumberToShortenCode_UniquenessOnSmallRange(t *testing.T) {
	// verify no collision within a small range (0..9999)
	seen := make(map[string]struct{}, 10000)
	for x := uint64(0); x < 10000; x++ {
		code, k := EncodeNumberToShortenCode(x)
		require.Equal(t, 5, k)
		_, exists := seen[code]
		assert.Falsef(t, exists, "duplicate code for x=%d: %q", x, code)
		seen[code] = struct{}{}
	}
}

func TestEncodeNumberToShortenCode_MonotonicRegionsLength(t *testing.T) {
	// ensure when the input go cross the border, getKForValue will change accordingly (5->6->7->8)
	cases := map[uint64]uint64{
		Pow62_5 - 1: Pow62_5,
		Pow62_6 - 1: Pow62_6,
		Pow62_7 - 1: Pow62_7,
	}
	for x1, x2 := range cases {
		code1, k1 := EncodeNumberToShortenCode(x1)
		code2, k2 := EncodeNumberToShortenCode(x2)
		assert.Equalf(t, getKForValue(x1), k1, "x1=%d", x1)
		assert.Lenf(t, code1, k1, "x1=%d", x1)
		assert.Equalf(t, getKForValue(x2), k2, "x2=%d", x2)
		assert.Lenf(t, code2, k2, "x2=%d", x2)

		// the length of the second code should be larger by 1
		assert.Equalf(t, k1+1, k2, "%d + 1 = %d", k1, k2)
	}
}

func TestEncodeFixedBase62_BasicAndMax(t *testing.T) {
	// basic
	assert.Equal(t, "00000", encodeFixedBase62(0, 5))
	assert.Equal(t, "00001", encodeFixedBase62(1, 5))

	// 61 -> 'z'
	assert.Equal(t, "0000z", encodeFixedBase62(61, 5))

	// 62 -> 1*62 + 0 => "00010"
	assert.Equal(t, "00010", encodeFixedBase62(62, 5))

	// max values
	assert.Equal(t, "zzzzz", encodeFixedBase62(Pow62_5-1, 5))
	assert.Equal(t, "zzzzzz", encodeFixedBase62(Pow62_6-1, 6))
	assert.Equal(t, "zzzzzzz", encodeFixedBase62(Pow62_7-1, 7))
	assert.Equal(t, "zzzzzzzz", encodeFixedBase62(Pow62_8-1, 8))
}

func TestRotateRightBy1(t *testing.T) {
	cases := map[string]string{
		"":      "",
		"a":     "a",
		"ab":    "ba",
		"abc":   "cab",
		"00001": "10000",
		"A0z9":  "9A0z",
	}
	for in, want := range cases {
		got := rotateRightBy1(in)
		assert.Equalf(t, want, got, "input=%q", in)
	}
}

func TestPow62(t *testing.T) {
	// supported k
	assert.Equal(t, Pow62_5, pow62(5))
	assert.Equal(t, Pow62_6, pow62(6))
	assert.Equal(t, Pow62_7, pow62(7))
	assert.Equal(t, Pow62_8, pow62(8))

	// unsupported k falls back to 62^8
	assert.Equal(t, Pow62_8, pow62(4))
	assert.Equal(t, Pow62_8, pow62(9))
	assert.Equal(t, Pow62_8, pow62(-1))
}

func TestGetKForValue(t *testing.T) {
	cases := []struct {
		x    uint64
		want int
	}{
		{0, 5},
		{Pow62_5 - 1, 5},
		{Pow62_5, 6},
		{Pow62_6 - 1, 6},
		{Pow62_6, 7},
		{Pow62_7 - 1, 7},
		{Pow62_7, 8},
		{Pow62_8 - 1, 8},
		{Pow62_8, 8},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, getKForValue(tc.x), "x=%d", tc.x)
	}
}
