package main

import (
	"net/http"
	"testing"
)

func TestStatus204NoContent(t *testing.T) {
	verifyResponse := func(t *testing.T, resp *http.Response) {
		t.Helper()
		if got, want := resp.StatusCode, http.StatusNoContent; got != want {
			t.Errorf("status code mismatch, got=%d, want=%d", got, want)
		}
		if got, want := resp.Header.Get("Content-Length"), ""; got != want {
			t.Errorf("content-length must not exist in response with status 204")
		}
	}

	c := NewTestClient(t, baseURL)

	res := c.Get("/status204")
	verifyResponse(t, res)

	res = c.Get("/status204")
	verifyResponse(t, res)
}
