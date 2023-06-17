package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestStatus200NoContent(t *testing.T) {
	verifyResponse := func(t *testing.T, resp *http.Response) {
		t.Helper()
		if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Errorf("status code mismatch, got=%d, want=%d", got, want)
		}
		if got, want := resp.Header.Get("Content-Length"), ""; got == want {
			t.Errorf("content-length expected to be set in this test's response")
		}
	}

	c := newTestClient(t)
	urlPath := fmt.Sprintf("/status?s=200&s-maxage=2&scenario=%s", NewScenarioID())

	res := c.Get(urlPath)
	verifyResponse(t, res)

	res = c.Get(urlPath)
	verifyResponse(t, res)
}
