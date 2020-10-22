package hash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSHA256Hash(t *testing.T) {
	h, err := SHA256Hash("salt", "1234")
	require.NoError(t, err)
	require.Equal(t, "ea32961dbd579ef5697c367f9267921ee07f14d77fb2d4fb9500d4221d615695", h)
	require.Len(t, h, 64)

	h2, err := SHA256Hash("new salt", "1234")
	require.NoError(t, err)
	require.NotEqual(t, h, h2)
}
