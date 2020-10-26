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

func Test_http_echo(t *testing.T) {
	client := NewHttpClient("", WithDebug())
	in := &PostIn{
		DummyField: "a123456",
	}
	out := &PostIn{}
	err := client.Post("http://127.0.0.1:3000/notif", in, out)
	require.NoError(t, err)
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

func Test_debug_unreachable_host(t *testing.T) {
	client := NewHttpClient("https://postman-echo.com/post", WithDebug())
	in := &PostIn{
		DummyField: "a123456",
	}
	out := &PostIn{}
	err := client.Post("http://192.168.1.189", in, out)
	require.Error(t, err)
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
