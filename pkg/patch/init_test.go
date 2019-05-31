package patch

import (
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	meta := InitMeta{
		Ports:              "80,443",
		EgressIgnoredPorts: "22",
		ContainerImage:     "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:v2",
		IgnoredIPs:         "169.254.169.254",
	}

	init, err := renderInit(meta)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(init, "80,443") {
		t.Errorf("Ports not found")
	}
}
