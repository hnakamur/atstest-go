package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hnakamur/atstest"
)

func main() {
	tsPort := flag.Int("ts-port", 8080, "trafficserver port")
	debug := flag.Bool("debug", false, "enable debug")
	flag.Parse()

	if err := run(*tsPort, *debug); err != nil {
		log.Fatal(err)
	}
}

func run(tsPort int, debug bool) (err error) {
	verifyResponse := func(resp *http.Response) {
		if got, want := resp.StatusCode, http.StatusNoContent; got != want {
			log.Printf("status code mismatch, got=%d, want=%d", got, want)
		}
		if got, want := resp.Header.Get("Content-Length"), ""; got != want {
			log.Printf("content-length must not exist in response with status 204")
		}
	}

	baseURL := fmt.Sprintf("http://localhost:%d", tsPort)
	c := atstest.NewTestClient(&atstest.LogTB{}, baseURL, debug)
	urlPath := fmt.Sprintf("/status?s=204&s-maxage=2&scenario=%s", atstest.NewScenarioID())

	log.Print("sending first request")
	res := c.Get(urlPath)
	verifyResponse(res)

	log.Print("sending second request")
	res = c.Get(urlPath)
	verifyResponse(res)

	return nil
}
