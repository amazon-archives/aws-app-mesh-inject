package config

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var TestConfig = Config{
	"mesh_name",
	"mesh_region",
	"mesh_log_level",
	false,
}

func init() {
	os.Setenv("APPMESH_NAME", TestConfig.Name)
	os.Setenv("APPMESH_REGION", TestConfig.Region)
	os.Setenv("APPMESH_LOG_LEVEL", TestConfig.LogLevel)
}

func TestInit(t *testing.T) {
	flag.Parse()
	fmt.Println(DefaultConfig.LogLevel)
}
