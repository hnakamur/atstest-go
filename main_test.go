package atstest

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code, err := doTestMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

var baseURL string
var debug bool

func doTestMain(m *testing.M) (code int, err error) {
	tsRoot := flag.String("ts-root", "/etc/trafficserver", "trafficserver root directory")
	tsFilename := flag.String("ts-filename", "traffic_server", "trafficserver filename or full path")
	tsUser := flag.String("ts-user", "trafficserver", "user name to run traffic_server")
	tsPort := flag.Int("ts-port", 8080, "trafficserver port")
	originPort := flag.Int("origin-port", 8880, "origin server port")
	waitBeforeTest := flag.Duration("wait-before-test", 0, "wait interval before test")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.Parse()

	baseURL = fmt.Sprintf("http://localhost:%d", *tsPort)

	origServer := NewOriginServer(fmt.Sprintf(":%d", *originPort))
	origErrC := make(chan error)
	go func() {
		origErrC <- origServer.ListenAndServe()
	}()
	defer func() {
		err2 := origServer.Shutdown(context.Background())
		err3 := <-origErrC
		if err3 == http.ErrServerClosed {
			err3 = nil
		}
		err = joinErrors(err, err2, err3)
	}()

	tsRunner := NewTrafficServerRunner(*tsRoot, *tsFilename, *tsUser, *tsPort, *originPort)
	if err := tsRunner.ModifyConfigFiles(); err != nil {
		return 0, err
	}
	if err := tsRunner.Start(); err != nil {
		return 0, err
	}
	defer func() {
		err2 := tsRunner.Stop()
		err = joinErrors(err, err2)
	}()

	if *waitBeforeTest > 0 {
		fmt.Fprintf(os.Stderr, "started origin and traffic_server, wait %s\n", *waitBeforeTest)
		time.Sleep(*waitBeforeTest)
	}

	return m.Run(), nil
}

func joinErrors(errs ...error) error {
	n := 0
	var err2 error
	for _, err := range errs {
		if err != nil {
			err2 = err
			n++
		}
	}
	switch n {
	case 0:
		return nil
	case 1:
		return err2
	default:
		return errors.Join(errs...)
	}
}

func newTestClient(t *testing.T) *TestClient {
	return NewTestClient(t, baseURL, debug)
}
