package patch

import (
	"strings"
	"testing"
)

func TestGeneratePatch(t *testing.T) {
	meta := Meta{
		AppendImagePullSecret: false,
		HasImagePullSecret:    false,
		AppendSidecar:         true,
		AppendInit:            false,
		Init: InitMeta{
			Ports:          "80,443",
			ContainerImage: "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest",
			IgnoredIPs:     "169.254.169.254",
		},
		Sidecar: SidecarMeta{
			MeshName:        "global",
			VirtualNodeName: "podinfo",
			Region:          "us-west-2",
			LogLevel:        "debug",
			ContainerImage:  "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
			CpuRequests:     "10m",
			MemoryRequests:  "32Mi",
		},
	}

	patch, err := GeneratePatch(meta)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(patch), "aws-appmesh-proxy-route-manager:latest") {
		t.Errorf("Init container image not found")
	}

	if !strings.Contains(string(patch), "aws-appmesh-envoy:latest") {
		t.Errorf("Sidecar container image not found")
	}
}
