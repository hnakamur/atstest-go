package atstest

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type TrafficserverRunner struct {
	tsRoot       string
	filename     string
	username     string
	port         int
	originPort   int
	majorVersion int
	cmd          *exec.Cmd
}

func NewTrafficServerRunner(tsRoot, filename, username string, port, originPort int) *TrafficserverRunner {
	r := &TrafficserverRunner{
		tsRoot:     tsRoot,
		filename:   filename,
		username:   username,
		port:       port,
		originPort: originPort,
	}
	r.majorVersion = r.GetMajorVersion()
	return r
}

func (r *TrafficserverRunner) Start() error {
	cmd := exec.Command(r.filename, "-p", strconv.Itoa(r.port))

	// overwrite or append TS_ROOT environment variable
	envs := make([]string, 0, len(os.Environ()))
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TS_ROOT=") {
			continue
		}
		envs = append(envs, env)
	}
	envs = append(envs, fmt.Sprintf("TS_ROOT=%s", r.tsRoot))
	cmd.Env = envs

	if err := cmd.Start(); err != nil {
		return err
	}
	r.cmd = cmd
	// log.Print("started trafficserver")
	if err := r.waitForStarted(); err != nil {
		return err
	}
	return nil
}

func (r *TrafficserverRunner) ModifyConfigFiles() error {
	if err := r.modifyRecordsYAMLorConfig(); err != nil {
		return err
	}
	if err := r.writeRemapConfig(); err != nil {
		return err
	}
	return nil
}

func (r *TrafficserverRunner) modifyRecordsYAMLorConfig() error {
	if r.majorVersion >= 10 {
		return r.modifyRecordsYAML()
	}
	return r.modifyRecordsConfig()
}

func (r *TrafficserverRunner) modifyRecordsYAML() error {
	// add or modify admin.user_id config

	configFilename := filepath.Join(r.tsRoot, "records.yaml")
	yamlBytes, err := os.ReadFile(configFilename)
	if err != nil {
		return err
	}

	var config map[string]any
	if err := yaml.Unmarshal(yamlBytes, &config); err != nil {
		return err
	}

	if config["ts"] == nil {
		config["ts"] = make(map[string]any)
	}

	cfgTS := config["ts"].(map[string]any)
	if cfgTS["admin"] == nil {
		cfgTS["admin"] = make(map[string]any)
	}

	cfgAdmin := cfgTS["admin"].(map[string]any)
	cfgAdmin["user_id"] = r.username

	yamlBytes2, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configFilename, yamlBytes2, 0o644); err != nil {
		return err
	}

	// Rename old config file if exists since traffic_server refuse to start.
	oldConfigFilename := filepath.Join(r.tsRoot, "records.config")
	if _, err := os.Stat(oldConfigFilename); err == nil {
		if err := os.Rename(oldConfigFilename, oldConfigFilename+".bak"); err != nil {
			return err
		}
	}

	return nil
}

func (r *TrafficserverRunner) modifyRecordsConfig() error {
	configFilename := filepath.Join(r.tsRoot, "records.config")
	configBytes, err := os.ReadFile(configFilename)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		configBytes, err = os.ReadFile(configFilename + ".bak")
		if err != nil {
			return err
		}
	}

	adminUserIDRe, err := regexp.Compile(`^CONFIG[ \t]+proxy\.config\.admin\.user_id[ \t]+STRING[ \t]+([^ ]+)`)
	if err != nil {
		return err
	}
	alreadyConfiguredAsNeeded := false
	modified := false
	var lines []string
	lineToAdd := fmt.Sprintf("CONFIG proxy.config.admin.user_id STRING %s", r.username)
	scanner := bufio.NewScanner(bytes.NewReader(configBytes))
	for scanner.Scan() {
		line := scanner.Text()
		m := adminUserIDRe.FindStringSubmatch(line)
		if len(m) == 2 {
			if m[1] == r.username {
				alreadyConfiguredAsNeeded = true
			} else {
				alreadyConfiguredAsNeeded = false
				line = lineToAdd
				modified = true
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !alreadyConfiguredAsNeeded && !modified {
		lines = append(lines, lineToAdd)
		modified = true
	}

	if modified {
		var b bytes.Buffer
		for _, line := range lines {
			fmt.Fprintf(&b, "%s\n", line)
		}
		if err := os.WriteFile(configFilename, b.Bytes(), 0o666); err != nil {
			return err
		}
	}

	return nil
}

func (r *TrafficserverRunner) writeRemapConfig() error {
	filename := filepath.Join(r.tsRoot, "remap.config")
	content := fmt.Sprintf("map / http://localhost:%d\n", r.originPort)
	if err := os.WriteFile(filename, []byte(content), 0o666); err != nil {
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
	for i := 0; i < 3; i++ {
		err = httpGet(baseURL)
		// log.Printf("wait for trafficserver, err=%v", err)
		if err == nil {
			return nil
		}
		// log.Print("sleep 1 second")
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

func (r *TrafficserverRunner) GetMajorVersion() int {
	cmd := exec.Command(r.filename, "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var v int
	n, err := fmt.Sscanf(string(output), "Traffic Server %d", &v)
	if err != nil {
		log.Fatal(err)
	}
	if n != 1 {
		log.Fatalf("cannot detect traffic_server major version: output=%q", string(output))
	}
	return v
}
