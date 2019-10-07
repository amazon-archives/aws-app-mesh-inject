package patch

import (
	"encoding/json"
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	meta := InitMeta{
		Ports:              "80,443",
		EgressIgnoredPorts: "22",
		ContainerImage:     "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:v2",
		IgnoredIPs:         "169.254.169.254",
		CpuRequests:        "100m",
		MemoryRequests:     "128Mi",
	}

	init, err := renderInit(meta)
	if err != nil {
		t.Fatal(err)
	}

	if !json.Valid([]byte(init)) {
		t.Fatal("invalid json")
	}

	if !strings.Contains(init, "80,443") {
		t.Errorf("Ports not found")
	}

	if !strings.Contains(init, "100m") {
		t.Errorf("CPU request not found")
	}

	if !strings.Contains(init, "128Mi") {
		t.Errorf("Memory request not found")
	}
}
