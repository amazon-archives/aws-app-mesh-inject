package patch

import (
	"strings"
	"testing"
)

func Test_Sidecar(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:        "debug",
		Region:          "us-west-2",
		VirtualNodeName: "podinfo",
		MeshName:        "global",
		ContainerImage:  "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:     "100m",
		MemoryRequests:  "128Mi",
	}

	sidecar, err := renderSidecar(meta)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sidecar, "mesh/global/virtualNode/podinfo") {
		t.Errorf("Virtual node not found")
	}
}
