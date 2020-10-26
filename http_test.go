package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_debug_http(t *testing.T) {
	client := NewHttpClient("http://www.google.com", WithDebug())
	var out interface{}
	err := client.Get("http://www.google.com", out)
	require.NoError(t, err)
}

type PostIn struct {
	DummyField string
}

func Test_debug_http_postman(t *testing.T) {
	client := NewHttpClient("https://postman-echo.com/post", WithDebug())
	in := &PostIn{
		DummyField: "a123456",
	}
	out := &PostIn{}
	err := client.Post("https://postman-echo.com/post", in, out)
	require.NoError(t, err)
}

func Test_debug_http_postman_without_debug(t *testing.T) {
	client := NewHttpClient("https://postman-echo.com/post")
	in := &PostIn{
		DummyField: "a123456",
	}
	out := &PostIn{}
	err := client.Post("https://postman-echo.com/post", in, out)
	require.NoError(t, err)
}
