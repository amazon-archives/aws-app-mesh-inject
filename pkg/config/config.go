package config

type Config struct {
	// HTTP Server settings
	Port    int
	TlsCert string
	TlsKey  string

	// Injetion Settings
	InjectDefault bool

	// If enabled, an fsGroup: 1337 will be injected in the absence of it within pod securityContext
	// see https://github.com/aws/amazon-eks-pod-identity-webhook/issues/8 for more details
	EnableIAMForServiceAccounts bool

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
