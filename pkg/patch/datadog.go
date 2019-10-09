package patch

import (
	"bufio"
	"bytes"
	"text/template"
)

// Creating a template to avoid relying on an extra ConfigMap
const datadogTemplate = `
tracing:
  http:
    name: envoy.tracers.datadog
    config:
      collector_cluster: datadog_agent
      service_name: envoy
static_resources:
  clusters:
  - name: datadog_agent
    connect_timeout: 1s
    type: strict_dns
    lb_policy: round_robin
    load_assignment:
      cluster_name: datadog_agent
      endpoints:
      - lb_endpoints:
        - endpoint:
           address:
            socket_address:
             address: {{ .Address }}
             port_value: {{ .Port }}
`

const injectDatadogTemplate = `
{
  "command": [
    "sh",
    "-c",
    "cat <<EOF >> /tmp/envoy/envoyconf.yaml{{ .Config }}EOF\n\ncat /tmp/envoy/envoyconf.yaml\n"
  ],
  "image": "busybox",
  "imagePullPolicy": "IfNotPresent",
  "name": "inject-datadog-config",
  "volumeMounts": [
  	{
      "mountPath": "/tmp/envoy",
      "name": "config"
    }
  ],
  "resources": {
    "limits": {
      "cpu": "100m",
      "memory": "64Mi"
    },
    "requests": {
      "cpu": "10m",
      "memory": "32Mi"
    }
  }
}
`

// renderDatadogInitContainer creates a container named inject-datadog-config
// that writes the Envoy config in an empty dir volume
// the same volume is mounted in the Envoy container at /tmp/envoy/
// when Envoy starts it will load the tracing config
func renderDatadogInitContainer(address string, port string) (string, error) {
	tmpl, err := template.New("datadog").Parse(datadogTemplate)
	if err != nil {
		return "", err
	}

	confModel := struct {
		Address string
		Port    string
	}{
		address,
		port,
	}

	var confData bytes.Buffer
	confWriter := bufio.NewWriter(&confData)
	if err := tmpl.Execute(confWriter, confModel); err != nil {
		return "", err
	}
	err = confWriter.Flush()
	if err != nil {
		return "", err
	}

	config, err := escapeYaml(confData.String())
	if err != nil {
		return "", err
	}

	tmplInit, err := template.New("initConfig").Parse(injectDatadogTemplate)
	if err != nil {
		return "", err
	}

	initModel := struct {
		Config string
	}{
		config,
	}

	var initData bytes.Buffer
	initWriter := bufio.NewWriter(&initData)
	if err := tmplInit.Execute(initWriter, initModel); err != nil {
		return "", err
	}
	err = initWriter.Flush()
	if err != nil {
		return "", err
	}

	return initData.String(), nil
}

// shared volume between the init container and Envoy
func renderDatadogConfigVolume() string {
	return configVolume
}