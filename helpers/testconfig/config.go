package testconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v2"
)

type TestConfig struct {
	OtgHost          string   `yaml:"otg_host,omitempty"`
	OtgPorts         []string `yaml:"otg_ports,omitempty"`
	OtgSpeed         string   `yaml:"otg_speed,omitempty"`
	OtgIterations    int      `yaml:"otg_iterations,omitempty"`
	OtgCaptureCheck  bool     `yaml:"otg_capture_check,omitempty"`
	OtgGrpcTransport bool     `yaml:"otg_grpc_transport,omitempty"`
}

func testConfigPath() (string, error) {
	_, src, _, _ := runtime.Caller(0)
	for {
		src = filepath.Dir(src)
		if src == filepath.Dir(src) {
			return "", fmt.Errorf("path exhausted")
		}
		files, err := ioutil.ReadDir(src)
		if err != nil {
			return "", err
		}
		for _, f := range files {
			if f.Name() == "versions.yaml" {
				return path.Join(src, "test-config.yaml"), nil
			}
		}
	}
}

func NewTestConfig(t *testing.T) *TestConfig {
	tc := TestConfig{
		OtgHost:          "https://localhost",
		OtgPorts:         []string{"localhost:5555", "localhost:5556"},
		OtgSpeed:         "speed_1_gbps",
		OtgIterations:    100,
		OtgCaptureCheck:  true,
		OtgGrpcTransport: false,
	}

	path, err := testConfigPath()
	if err != nil {
		t.Fatalf("Could not get test config path: %v\n", err)
	}

	if err := tc.Load(t, path); err != nil {
		t.Fatalf("Could not load test config: %v", err)
	}

	return &tc
}

func (tc *TestConfig) Load(t *testing.T, path string) error {
	b, err := os.ReadFile(path)

	if err != nil {
		t.Logf("Using default test configuration because: %v\n", err)
		return nil
	}

	t.Logf("Loading test config from %s\n", path)
	if err := yaml.Unmarshal(b, tc); err != nil {
		return fmt.Errorf("could not unmarshal %s: %v", path, err)
	}

	return nil
}
