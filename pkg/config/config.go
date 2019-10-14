package config

type Config struct {
	// HTTP Server settings
	Port    int
	TlsCert string
	TlsKey  string

	// Sidecar settings
	SidecarImage  string
	SidecarCpu    string
	SidecarMemory string
	MeshName      string
	Region        string
	Preview       bool
	LogLevel      string
	EcrSecret     bool

	// Init container settings
	InitImage  string
	IgnoredIPs string

	// Observability settings
	InjectXraySidecar           bool
	EnableStatsTags             bool
	EnableStatsD                bool
	InjectStatsDExporterSidecar bool
	EnableJaegerTracing         bool
	JaegerAddress               string
	JaegerPort                  string
	EnableDatadogTracing        bool
	DatadogAddress              string
	DatadogPort                 string
}
