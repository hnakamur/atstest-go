package main

import (
	"errors"
	"fmt"
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
	c, err := NewClient(baseURL)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		_, err = c.Get("/")
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
