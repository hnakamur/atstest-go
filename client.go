package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL: u,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (c *Client) Get(urlPath string) (*http.Response, error) {
	u := c.baseURL.JoinPath(urlPath)
	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	if err := reloadResponseBody(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func reloadResponseBody(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return nil
}
