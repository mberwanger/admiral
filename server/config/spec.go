package config

import (
	"fmt"
	"time"
	
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Server    Server    `yaml:"server"`
	Endpoints Endpoints `yaml:"endpoints"`
	Services  Services  `yaml:"services"`
}

type Server struct {
	Listener             Listener   `yaml:"listener"`
	Timeouts             Timeouts   `yaml:"timeouts"`
	Logger               *Logger    `yaml:"logger"`
	AccessLog            *AccessLog `yaml:"access_log"`
	Stats                *Stats     `yaml:"stats"`
	EnablePprof          bool       `yaml:"enable_pprof"`
	MaxResponseSizeBytes int        `yaml:"max_response_size_bytes"`
}

type Listener struct {
	Address string `yaml:"address" validate:"ip"`
	Port    int    `yaml:"port" validate:"required,min=1,max=65535"`
}

type Timeouts struct {
	Default   time.Duration   `yaml:"default"`
	Overrides []TimeoutsEntry `yaml:"overrides"`
}

type TimeoutsEntry struct {
	Service string        `yaml:"service"`
	Method  string        `yaml:"method"`
	Timeout time.Duration `yaml:"timeout"`
}

type Logger struct {
	Level     zapcore.Level `yaml:"level"`
	Namespace string        `yaml:"namespace"`
	Pretty    bool          `yaml:"pretty"`
}

type AccessLog struct {
	StatusCodeFilters []int `yaml:"status_code_filters"`
}

type Stats struct {
	FlushInterval  time.Duration   `yaml:"flush_interval"`
	GoRuntimeStats *GoRuntimeStats `yaml:"go_runtime_stats"`
	Prefix         string          `yaml:"prefix"`
	ReporterType   ReporterType    `yaml:"reporter_type"`
}

type GoRuntimeStats struct {
	CollectionInterval *time.Duration `yaml:"collection_interval" validate:"required,min=1s"`
}

type ReporterType string

const (
	ReporterTypeNull       ReporterType = "null"
	ReporterTypeLog        ReporterType = "log"
	ReporterTypePrometheus ReporterType = "prometheus"
)

func (r *ReporterType) String() string {
	return string(*r)
}

func (r *ReporterType) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}

func (r *ReporterType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	switch str {
	case string(ReporterTypeNull):
		*r = ReporterTypeNull
	case string(ReporterTypeLog):
		*r = ReporterTypeLog
	case string(ReporterTypePrometheus):
		*r = ReporterTypePrometheus
	default:
		return fmt.Errorf("invalid reporter type: %q", str)
	}

	return nil
}

func (r *ReporterType) Validate() error {
	switch *r {
	case ReporterTypeNull, ReporterTypeLog, ReporterTypePrometheus:
		return nil
	default:
		return fmt.Errorf("invalid reporter type: %s", *r)
	}
}

type Endpoints struct {
}

type Services struct {
	Database *Database `yaml:"database"`
}

type Database struct {
	Host         string  `yaml:"host"`
	Port         int     `yaml:"port"`
	DatabaseName string  `yaml:"database_name"`
	User         string  `yaml:"user"`
	Password     string  `yaml:"password"`
	SSLMode      SSLMode `yaml:"ssl_mode"`
}

// SSLMode represents different SSL connection modes.
type SSLMode int

// Constants defining the possible SSL modes.
const (
	SSLModeUnspecified SSLMode = 0
	SSLModeDisable     SSLMode = 1
	SSLModeAllow       SSLMode = 2
	SSLModePrefer      SSLMode = 3
	SSLModeRequire     SSLMode = 4
	SSLModeVerifyCA    SSLMode = 5
	SSLModeVerifyFull  SSLMode = 6
)

// SSLModeName maps SSLMode values to their string representations.
var SSLModeName = map[int]string{
	0: "unspecified",
	1: "disable",
	2: "allow",
	3: "prefer",
	4: "require",
	5: "verify_ca",
	6: "verify_full",
}

// SSLModeValue maps string representations to SSLMode values.
var SSLModeValue = map[string]int{
	"unspecified": 0,
	"disable":     1,
	"allow":       2,
	"prefer":      3,
	"require":     4,
	"verify_ca":   5,
	"verify_full": 6,
}

// String returns the string representation of the SSLMode.
func (s *SSLMode) String() string {
	if name, ok := SSLModeName[int(*s)]; ok {
		return name
	}
	return fmt.Sprintf("SSLMode(%d)", *s)
}

// MarshalYAML marshals the SSLMode to YAML as a string.
func (s *SSLMode) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

// UnmarshalYAML unmarshals a YAML string into an SSLMode.
func (s *SSLMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	if value, ok := SSLModeValue[str]; ok {
		*s = SSLMode(value)
		return nil
	}
	return fmt.Errorf("invalid SSLMode: %q", str)
}

// Validate checks if the SSLMode value is valid.
func (s *SSLMode) Validate() error {
	_, valid := SSLModeName[int(*s)]
	if !valid {
		return fmt.Errorf("invalid SSLMode: %d", *s)
	}
	return nil
}
