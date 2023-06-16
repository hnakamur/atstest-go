package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code, err := doTestMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

var baseURL string

func doTestMain(m *testing.M) (code int, err error) {
	tsFilename := flag.String("ts-filename", "traffic_server", "trafficserver filename or full path")
	tsPort := flag.Int("ts-port", 8080, "trafficserver port")
	flag.Parse()

	baseURL = fmt.Sprintf("http://localhost:%d", *tsPort)

	origServer := NewOriginServer(":8088")
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

	tsRunner := NewTrafficServerRunner(*tsFilename, *tsPort)
	if err := tsRunner.Start(); err != nil {
		return 0, err
	}
	defer func() {
		err2 := tsRunner.Stop()
		err = joinErrors(err, err2)
	}()

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
