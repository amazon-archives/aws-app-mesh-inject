package patch

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_Sidecar(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:        "debug",
		Region:          "us-west-2",
		Preview:         "0",
		VirtualNodeName: "podinfo",
		MeshName:        "global",
		ContainerImage:  "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:     "100m",
		MemoryRequests:  "128Mi",
	}

	checkSidecars(t, meta)
}

func Test_Sidecar_WithXray(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:          "debug",
		Region:            "us-west-2",
		Preview:           "0",
		VirtualNodeName:   "podinfo",
		MeshName:          "global",
		ContainerImage:    "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:       "100m",
		MemoryRequests:    "128Mi",
		InjectXraySidecar: true,
	}

	checkSidecars(t, meta)
}

func Test_Sidecar_WithStatsTags(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:        "debug",
		Region:          "us-west-2",
		Preview:         "0",
		VirtualNodeName: "podinfo",
		MeshName:        "global",
		ContainerImage:  "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:     "100m",
		MemoryRequests:  "128Mi",
		EnableStatsTags: true,
	}

	checkSidecars(t, meta)
}

func Test_Sidecar_WithStatsD(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:                    "debug",
		Region:                      "us-west-2",
		Preview:                     "0",
		VirtualNodeName:             "podinfo",
		MeshName:                    "global",
		ContainerImage:              "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:                 "100m",
		MemoryRequests:              "128Mi",
		EnableStatsTags:             true,
		EnableStatsD:                true,
		InjectStatsDExporterSidecar: true,
	}

	checkSidecars(t, meta)
}

func Test_Sidecar_WithDatadog(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:             "debug",
		Region:               "us-west-2",
		Preview:              "0",
		VirtualNodeName:      "podinfo",
		MeshName:             "global",
		ContainerImage:       "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:          "100m",
		MemoryRequests:       "128Mi",
		EnableDatadogTracing: true,
		DatadogAddress:       "datadog.appmesh-system",
		DatadogPort:          "8126",
	}

	checkSidecars(t, meta)
}

func Test_Sidecar_WithJaeger(t *testing.T) {
	meta := SidecarMeta{
		LogLevel:            "debug",
		Region:              "us-west-2",
		Preview:             "0",
		VirtualNodeName:     "podinfo",
		MeshName:            "global",
		ContainerImage:      "111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:latest",
		CpuRequests:         "100m",
		MemoryRequests:      "128Mi",
		EnableJaegerTracing: true,
		JaegerAddress:       "appmesh-jaeger.appmesh-system",
		JaegerPort:          "9411",
	}

	checkSidecars(t, meta)
}

func checkSidecars(t *testing.T, meta SidecarMeta) {
	var err error

	sidecars, err := renderSidecars(meta)
	if err != nil {
		t.Fatal(err)
	}

	for _, sidecar := range sidecars {
		var v interface{}
		err = json.Unmarshal([]byte(sidecar), &v)
		if err != nil {
			t.Fatal(err)
		}
		cm := v.(map[string]interface{})
		switch cm["name"] {
		case "envoy":
			checkEnvoy(t, cm, meta)
		case "xray-daemon":
			checkXrayDaemon(t, cm, meta)
		case "statsd-exporter":
			checkStatsDExporter(t, cm, meta)
		default:
			t.Errorf("Unexpected container found with name %s", cm["name"])
		}
	}
}

func checkEnvoy(t *testing.T, m map[string]interface{}, meta SidecarMeta) {
	expectedEnvs := map[string]string{
		"APPMESH_VIRTUAL_NODE_NAME": fmt.Sprintf("mesh/%s/virtualNode/%s", meta.MeshName, meta.VirtualNodeName),
		"AWS_REGION":                meta.Region,
		"ENVOY_LOG_LEVEL":           meta.LogLevel,
		"APPMESH_PREVIEW":           meta.Preview,
	}

	if meta.EnableJaegerTracing || meta.EnableDatadogTracing {
		expectedEnvs["ENVOY_STATS_CONFIG_FILE"] = "/tmp/envoy/envoyconf.yaml"

		mounts := m["volumeMounts"].([]interface{})
		if len(mounts) < 1 {
			t.Errorf("no volume mounts found")
		}

		mount := mounts[0].(map[string]interface{})
		mountName := mount["name"].(string)
		expectedMountName := "config"
		if mountName != expectedMountName {
			t.Errorf("volume mount name is set to %s instead of %s", mountName, expectedMountName)
		}

		mountPath := mount["mountPath"].(string)
		expectedMountPath := "/tmp/envoy"
		if mountPath != expectedMountPath {
			t.Errorf("volume mount path is set to %s instead of %s", mountPath, expectedMountPath)
		}
	}

	if meta.InjectXraySidecar {
		expectedEnvs["ENABLE_ENVOY_XRAY_TRACING"] = "1"
	}

	if meta.EnableStatsTags {
		expectedEnvs["ENABLE_ENVOY_STATS_TAGS"] = "1"
	}

	if meta.EnableStatsD {
		expectedEnvs["ENABLE_ENVOY_DOG_STATSD"] = "1"
	}

	if m["image"] != meta.ContainerImage {
		t.Errorf("Envoy container image is not set to %s", meta.ContainerImage)
	}

	envs := m["env"].([]interface{})
	for _, u := range envs {
		item := u.(map[string]interface{})
		name := item["name"].(string)
		if expected, ok := expectedEnvs[name]; ok {
			val := item["value"].(string)
			if val != expected {
				t.Errorf("%s env is set %s instead of %s", name, val, expected)
			} else {
				delete(expectedEnvs, name)
			}
		}
	}

	for k := range expectedEnvs {
		t.Errorf("%s env is not set", k)
	}
}

func checkXrayDaemon(t *testing.T, m map[string]interface{}, meta SidecarMeta) {
	if !meta.InjectXraySidecar {
		t.Errorf("Xray daemon is added when InjectXraySidecar is false")
	}

	if m["image"] != "amazon/aws-xray-daemon" {
		t.Errorf("Xray daemon container image is not set to amazon/aws-xray-daemon")
	}
}

func checkStatsDExporter(t *testing.T, m map[string]interface{}, meta SidecarMeta) {
	if !meta.InjectStatsDExporterSidecar {
		t.Errorf("StatsD exporter sidecar is added when InjectStatsDExporterSidecar is false")
	}

	if m["image"] != "maddox/statsd-exporter" {
		t.Errorf("StatsD exporter container image is not set to maddox/statsd-exporter")
	}
}
