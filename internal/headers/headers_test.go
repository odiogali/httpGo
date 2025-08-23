package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with whitespace before and after
	headers = NewHeaders()
	data = []byte("          Host:      localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 42, n)
	assert.False(t, done)

	// Test: Valid single header with spacing and special character
	headers = NewHeaders()
	data = []byte("    Content-Length: 139       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	// Test: Valid multiple headers
	// headers = NewHeaders()
	// data = []byte("Host: localhost:42069\r\nAccept-Language:     en-US,en;q=0.5\r\n\r\n")
	// n, done, err = headers.Parse(data)
	// require.NoError(t, err)
	// require.NotNil(t, headers)
	// assert.Equal(t, "localhost:42069", headers["Host"])
	// assert.Equal(t, "localhost:42069", headers["Accept-Language"])
	// //assert.Equal(t, 25, n)
	// assert.True(t, done)
}
