package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/hnakamur/atstest"
)

func main() {
	tsRoot := flag.String("ts-root", "/etc/trafficserver", "trafficserver root directory")
	tsFilename := flag.String("ts-filename", "traffic_server", "trafficserver filename or full path")
	tsUser := flag.String("ts-user", "trafficserver", "user name to run traffic_server")
	tsPort := flag.Int("ts-port", 8080, "trafficserver port")
	originPort := flag.Int("origin-port", 8880, "origin server port")
	flag.Parse()

	if err := run(*tsRoot, *tsFilename, *tsUser, *tsPort, *originPort); err != nil {
		log.Fatal(err)
	}
}

func run(tsRoot, tsFilename, tsUser string, tsPort, origPort int) (err error) {
	// origServer := atstest.NewOriginServer(fmt.Sprintf(":%d", origPort))
	// origErrC := make(chan error)
	// go func() {
	// 	origErrC <- origServer.ListenAndServe()
	// }()
	// defer func() {
	// 	err2 := origServer.Shutdown(context.Background())
	// 	err3 := <-origErrC
	// 	if err3 == http.ErrServerClosed {
	// 		err3 = nil
	// 	}
	// 	err = joinErrors(err, err2, err3)
	// }()

	tsRunner := atstest.NewTrafficServerRunner(tsRoot, tsFilename, tsUser, tsPort, origPort)
	v := tsRunner.GetMajorVersion()
	log.Printf("traffic_server major version=%d", v)

	if err := tsRunner.ModifyConfigFiles(); err != nil {
		return err
	}

	// if err := tsRunner.Start(); err != nil {
	// 	return err
	// }
	// defer func() {
	// 	err2 := tsRunner.Stop()
	// 	err = joinErrors(err, err2)
	// }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()

	return nil
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
