package config

import (
	"flag"
	"os"
)

var (
	DefaultConfig Config
)

func init() {
	flag.StringVar(&DefaultConfig.Name, "name", os.Getenv("APPMESH_NAME"), "AWS App Mesh name")
	flag.StringVar(&DefaultConfig.Region, "region", os.Getenv("APPMESH_REGION"), "AWS App Mesh region")
	flag.StringVar(&DefaultConfig.LogLevel, "log-level", os.Getenv("APPMESH_LOG_LEVEL"), "AWS App Mesh envoy log level")
	flag.BoolVar(&DefaultConfig.EcrSecret, "ecr-secret", false, "Inject AWS app mesh pull secrets")
}

type Config struct {
	Name      string
	Region    string
	LogLevel  string
	EcrSecret bool
}
