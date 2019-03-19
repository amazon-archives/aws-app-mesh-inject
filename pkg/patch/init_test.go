package patch

import (
	"fmt"
	"strings"
	"testing"
)

func Test_Init(t *testing.T) {
	meta := InitMeta{
		Ports:          "80,443",
		ContainerImage: "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest",
		IgnoredIPs:     "169.254.169.254",
	}

	sidecar, err := renderInit(meta)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(sidecar)

	if !strings.Contains(sidecar, "80,443") {
		t.Errorf("Ports not found")
	}
}
