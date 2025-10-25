package headers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:8080\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:8080", headers["host"])
	assert.Equal(t, 22, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("  Content-Type:   text/html  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html", headers["content-type"])
	assert.Equal(t, 31, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	headers["connection"] = "keep-alive"
	data = []byte("Host: example.com\r\nUser-Agent: curl/8.1\r\n\r\n")

	n1, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "example.com", headers["host"])

	fmt.Println(n1)
	fmt.Println(string(data[n1:]))
	n2, done, err := headers.Parse(data[n1:])
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "curl/8.1", headers["user-agent"])

	_, done, err = headers.Parse(data[n1+n2:])
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, "keep-alive", headers["connection"])

	// Test: Valid done (just CRLF)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 2, n)
	assert.Empty(t, headers)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:8080       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Capital letters in header keys
	headers = NewHeaders()
	data = []byte("X-Custom-Header: value\r\n\r\n")
	_, _, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "value", headers["x-custom-header"])

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:8080\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Starting header matches header to be parsed
	headers = NewHeaders()
	headers["set-person"] = "alice"
	data = []byte("Set-Person: bob\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 17, n)
	assert.Equal(t, "alice, bob", headers["set-person"])
	assert.False(t, done)
}
