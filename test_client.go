package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type TestClient struct {
	t          *testing.T
	baseURL    *url.URL
	httpClient *http.Client
}

func NewTestClient(t *testing.T, baseURL string) *TestClient {
	t.Helper()
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	return &TestClient{
		t:       t,
		baseURL: u,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *TestClient) Get(urlPath string) *http.Response {
	c.t.Helper()
	u := c.baseURL.JoinPath(urlPath)
	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		c.t.Fatal(err)
	}
	if err := reloadResponseBody(resp); err != nil {
		c.t.Fatal(err)
	}
	return resp
}

func reloadResponseBody(resp *http.Response) error {
	if resp.Body == nil {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return nil
}
