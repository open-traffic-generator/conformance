package testconfig

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"gopkg.in/yaml.v2"
)

type DutConfig struct {
	Name         string   `yaml:"name,omitempty"`
	Host         string   `yaml:"host,omitempty"`
	SshUsername  string   `yaml:"ssh_username,omitempty"`
	SshPassword  string   `yaml:"ssh_password,omitempty"`
	SshPort      int      `yaml:"ssh_port,omitempty"`
	GnmiUsername string   `yaml:"gnmi_username,omitempty"`
	GnmiPassword string   `yaml:"gnmi_password,omitempty"`
	GnmiPort     int      `yaml:"gnmi_port,omitempty"`
	Interfaces   []string `yaml:"interfaces,omitempty"`
}

type TestConfig struct {
	OtgHost          string                 `yaml:"otg_host,omitempty"`
	OtgPorts         []string               `yaml:"otg_ports,omitempty"`
	OtgSpeed         string                 `yaml:"otg_speed,omitempty"`
	OtgIterations    int                    `yaml:"otg_iterations,omitempty"`
	OtgCaptureCheck  bool                   `yaml:"otg_capture_check,omitempty"`
	OtgGrpcTransport bool                   `yaml:"otg_grpc_transport,omitempty"`
	DutConfigs       []DutConfig            `yaml:"dut_configs,omitempty"`
	OtgTestConst     map[string]interface{} `yaml:"otg_test_const,omitempty"`
}

func testConfigPath() (string, error) {
	_, src, _, _ := runtime.Caller(0)
	for {
		src = filepath.Dir(src)
		if src == filepath.Dir(src) {
			return "", fmt.Errorf("path exhausted")
		}
		files, err := os.ReadDir(src)
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
		OtgHost:          "https://localhost:8443",
		OtgPorts:         []string{"localhost:5555", "localhost:5556"},
		OtgSpeed:         "speed_1_gbps",
		DutConfigs:       []DutConfig{},
		OtgIterations:    100,
		OtgCaptureCheck:  true,
		OtgGrpcTransport: false,
	}

	path, err := testConfigPath()
	if err != nil {
		t.Fatalf("ERROR: Could not get test config path: %v\n", err)
	}

	if err := tc.Load(t, path); err != nil {
		t.Fatalf("ERROR: Could not load test config: %v\n", err)
	}

	if err := tc.FromEnv(); err != nil {
		t.Fatalf("ERROR: Could not load test config from env: %v\n", err)
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

func (tc *TestConfig) FromEnv() error {
	if s := os.Getenv("OTG_HOST"); s != "" {
		tc.OtgHost = s
	}
	if s := os.Getenv("OTG_ITERATIONS"); s != "" {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return fmt.Errorf("could not parse env OTG_ITERATIONS=%v: %v", v, err)
		}
		tc.OtgIterations = int(v)
	}
	if s := os.Getenv("OTG_CAPTURE_CHECK"); s != "" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("could not parse env OTG_CAPTURE_CHECK=%v: %v", v, err)
		}
		tc.OtgCaptureCheck = v
	}
	if s := os.Getenv("OTG_GRPC_TRANSPORT"); s != "" {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("could not parse env OTG_GRPC_TRANSPORT=%v: %v", v, err)
		}
		tc.OtgGrpcTransport = v
	}
	return nil
}

func (tc *TestConfig) PatchTestConst(t *testing.T, testConst map[string]interface{}) {
	if tc.OtgTestConst == nil {
		return
	}

	for k := range testConst {
		if v := tc.OtgTestConst[k]; v != nil {
			if testConst[k] != nil {
				testConst[k] = v
			}
		}
	}

	t.Log(tc.OtgTestConst)
	t.Log(testConst)
}
