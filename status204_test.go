package main

import (
	"fmt"
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

	c := newTestClient(t)
	urlPath := fmt.Sprintf("/status?s=204&s-maxage=2&scenario=%s", NewScenarioID())

	res := c.Get(urlPath)
	verifyResponse(t, res)

	res = c.Get(urlPath)
	verifyResponse(t, res)
}
