// Copyright 2016 Brad Rydzewski. All Rights Reserved.
// Use of this source code is governed by the open source Apache License, Version 2.0.

// Added bearer token and basic auth support with custom error handling

package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type HttpClient struct {
	client      *http.Client
	Base        string
	BearerToken string
	BasicAuth   struct {
		Username string
		Password string
	}
	debug bool
}

type HttpClientOption func(*HttpClient)

func buildClient(base string) *http.Client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if strings.HasPrefix(base, "https") {
		// Set up our own certificate pool
		tlsConfig := &tls.Config{RootCAs: x509.NewCertPool(), InsecureSkipVerify: true}
		transport.TLSClientConfig = tlsConfig
	}
	return &http.Client{Transport: transport}
}

func NewHttpClient(base string, opts ...HttpClientOption) *HttpClient {
	h := &HttpClient{client: buildClient(base), Base: base, debug: false}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func WithDebug() HttpClientOption {
	return func(h *HttpClient) {
		h.debug = true
	}
}

func (h *HttpClient) SetBearerToken(token string) {
	h.BearerToken = token
}

func (h *HttpClient) SetBasicAuth(username string, password string) {
	h.BasicAuth.Username = username
	h.BasicAuth.Password = password
}

//
// http request helper functions
//

// helper function for making an http GET request.
func (h *HttpClient) Get(rawurl string, out interface{}) error {
	return h.do(rawurl, "GET", nil, out)
}

// helper function for making an http POST request.
func (h *HttpClient) Post(rawurl string, in, out interface{}) error {
	return h.do(rawurl, "POST", in, out)
}

// helper function for making an http PUT request.
func (h *HttpClient) Put(rawurl string, in, out interface{}) error {
	return h.do(rawurl, "PUT", in, out)
}

// helper function for making an http PATCH request.
func (h *HttpClient) Patch(rawurl string, in, out interface{}) error {
	return h.do(rawurl, "PATCH", in, out)
}

// helper function for making an http DELETE request.
func (h *HttpClient) Delete(rawurl string, in, out interface{}) error {
	return h.do(rawurl, "DELETE", in, out)
}

// helper function to make an http request
func (h *HttpClient) do(rawurl, method string, in, out interface{}) error {
	resp, err := h.open(rawurl, method, in, out)
	if err != nil {
		if resp != nil {
			fmt.Printf("Response error status %s from %s at %s, error: %+v\n", resp.Status, resp.Request.URL, time.Now().Format(time.RFC3339), err)
		} else {
			fmt.Printf("Response error at %s, error: %+v\n", time.Now().Format(time.RFC3339), err)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		httpError := &HttpError{Status: resp.StatusCode}
		json.NewDecoder(resp.Body).Decode(httpError)
		if h.debug {
			jsonErr, _ := json.MarshalIndent(httpError, "", " ")
			fmt.Printf("Response http error status %s from %s at %s, httpError: %s\n", resp.Status, resp.Request.URL, time.Now().Format(time.RFC3339), jsonErr)
		}
		return httpError
	}

	if out != nil {
		reader := resp.Body
		rawData, err := ioutil.ReadAll(reader)
		if err != nil {
			fmt.Printf("Response body reader: %+v", err)
			return err
		}
		rawDataStr := string(rawData)
		if h.debug {
			fmt.Printf("Response status %s from %s at %s, data: %s\n", resp.Status, resp.Request.URL, time.Now().Format(time.RFC3339), rawDataStr)
		}
		err = json.NewDecoder(bytes.NewReader(rawData)).Decode(out)
		if err != nil {
			if h.debug {
				fmt.Printf("Response json decoding error data: %s err: %+v\n", rawDataStr, err)
			}
		}
	} else {
		if h.debug {
			fmt.Printf("Response status %s from %s at %s\n", resp.Status, resp.Request.URL, time.Now().Format(time.RFC3339))
		}
	}

	return nil
}

// helper function to open an http request
func (h *HttpClient) open(rawurl, method string, in, out interface{}) (*http.Response, error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	// creates a new http request to bitbucket.
	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return nil, err
	}

	if h.BearerToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.BearerToken))
	}

	if h.BasicAuth.Username != "" || h.BasicAuth.Password != "" {
		req.SetBasicAuth(h.BasicAuth.Username, h.BasicAuth.Password)
	}

	// if we are posting or putting data, we need to
	// write it to the body of the request.
	if in != nil {

		rc, ok := in.(io.ReadCloser)
		if ok {
			req.Body = rc
			req.Header.Set("Content-Type", "plain/text")
		} else {
			inJson, err := json.Marshal(in)
			if err != nil {
				return nil, err
			}

			if h.debug {
				fmt.Printf("Request %s to %s at %s, data: %s\n", method, uri.String(), time.Now().Format(time.RFC3339), inJson)
			}
			buf := bytes.NewBuffer(inJson)
			req.Body = ioutil.NopCloser(buf)

			req.ContentLength = int64(len(inJson))
			req.Header.Set("Content-Length", strconv.Itoa(len(inJson)))
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		if h.debug {
			fmt.Printf("Request %s to %s at %s\n", method, uri.String(), time.Now().Format(time.RFC3339))
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
