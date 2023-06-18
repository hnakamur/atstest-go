package atstest

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"
)

type TestClient struct {
	t          TBSub
	debug      bool
	baseURL    *url.URL
	httpClient *http.Client
}

func NewTestClient(t TBSub, baseURL string, debug bool) *TestClient {
	t.Helper()
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	return &TestClient{
		t:          t,
		debug:      debug,
		baseURL:    u,
		httpClient: &http.Client{
			// Timeout: 5 * time.Second,
		},
	}
}

func (c *TestClient) Get(urlPath string) *http.Response {
	c.t.Helper()
	u := c.baseURL.String() + urlPath
	resp, err := c.httpClient.Get(u)
	if err != nil {
		c.t.Fatal(err)
	}
	if err := reloadResponseBody(resp); err != nil {
		c.t.Fatal(err)
	}
	if c.debug {
		c.t.Logf("TestClient.Get, url=%s, resp=\n%+v", u, responseToString(resp))
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
	if err := resp.Body.Close(); err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return nil
}

func responseToString(resp *http.Response) string {
	// We cannot use http.Response.Write method here since it prevents
	// to write the Content-Length header in some cases.

	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%s %s\r\n", resp.Proto, resp.Status); err != nil {
		log.Fatal(err)
	}
	if err := resp.Header.Write(&b); err != nil {
		log.Fatal(err)
	}
	if _, err := b.Write([]byte("\r\n")); err != nil {
		log.Fatal(err)
	}
	bodyStart := b.Len()
	if _, err := io.Copy(&b, resp.Body); err != nil {
		log.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		log.Fatal(err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(b.Bytes()[bodyStart:]))
	return b.String()
}

func NewScenarioID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b[:])
}

type TBSub interface {
	Helper()
	Logf(format string, args ...any)
	Fatal(args ...any)
}

var _ TBSub = (*testing.T)(nil)
var _ TBSub = (*LogTB)(nil)

type LogTB struct{}

func (t *LogTB) Helper() {}

func (t *LogTB) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

func (t *LogTB) Fatal(args ...any) {
	log.Fatal(args...)
}
