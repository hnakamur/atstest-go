package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

type TrafficserverRunner struct {
	filename string
	port     int
	cmd      *exec.Cmd
}

func NewTrafficServerRunner(filename string, port int) *TrafficserverRunner {
	return &TrafficserverRunner{
		filename: filename,
		port:     port,
	}
}

func (r *TrafficserverRunner) Start() error {
	cmd := exec.Command(r.filename, "-p", strconv.Itoa(r.port))
	if err := cmd.Start(); err != nil {
		return err
	}
	r.cmd = cmd
	if err := r.waitForStarted(); err != nil {
		return err
	}
	return nil
}

func (r *TrafficserverRunner) waitForStarted() error {
	baseURL := fmt.Sprintf("http://localhost:%d", r.port)

	httpGet := func(baseURL string) error {
		resp, err := http.Get(baseURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return err
		}
		return nil
	}

	var err error
	for i := 0; i < 10; i++ {
		err = httpGet(baseURL)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return err
}

func (r *TrafficserverRunner) Stop() error {
	if err := r.cmd.Process.Kill(); err != nil {
		return err
	}
	if err := r.cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && !exitErr.Exited() &&
			exitErr.String() == "signal: killed" {
			return nil
		}
		return err
	}
	return nil
}
