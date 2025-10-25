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
	assert.Equal(t, "localhost:8080", headers["Host"])
	assert.Equal(t, 22, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("  Content-Type:   text/html  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html", headers["Content-Type"])
	assert.Equal(t, 31, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	headers["Connection"] = "keep-alive" // pre-existing
	data = []byte("Host: example.com\r\nUser-Agent: curl/8.1\r\n\r\n")

	n1, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "example.com", headers["Host"])

	fmt.Println(n1)
	fmt.Println(string(data[n1:]))
	n2, done, err := headers.Parse(data[n1:])
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "curl/8.1", headers["User-Agent"])

	_, done, err = headers.Parse(data[n1+n2:])
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, "keep-alive", headers["Connection"])

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
}
