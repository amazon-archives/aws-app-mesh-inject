package patch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"text/template"
)

const jaegerTemplate = `
tracing:
 http:
  name: envoy.zipkin
  typed_config:
   "@type": type.googleapis.com/envoy.config.trace.v2.ZipkinConfig
   collector_cluster: jaeger
   collector_endpoint: "/api/v1/spans"
   shared_span_context: false
static_resources:
  clusters:
  - name: jaeger
    connect_timeout: 1s
    type: strict_dns
    lb_policy: round_robin
    load_assignment:
      cluster_name: jaeger
      endpoints:
      - lb_endpoints:
        - endpoint:
           address:
            socket_address:
             address: {{ .Address }}
             port_value: {{ .Port }}
`

const injectJaegerTemplate = `
{
  "command": [
    "sh",
    "-c",
    "cat <<EOF >> /tmp/envoy/envoyconf.yaml{{ .Config }}EOF\n\ncat /tmp/envoy/envoyconf.yaml\n"
  ],
  "image": "busybox",
  "imagePullPolicy": "IfNotPresent",
  "name": "inject-jaeger-config",
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

const configVolume = `
{
  "name": "config",
  "emptyDir": {}
}
`

// renderJaegerInitContainer creates a container named inject-jaeger-config
// that writes the Envoy config in an empty dir volume
// the same volume is mounted in the Envoy container at /tmp/envoy/
// when Envoy starts it will load the tracing config
func renderJaegerInitContainer(address string, port string) (string, error) {
	tmpl, err := template.New("jaeger").Parse(jaegerTemplate)
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

	tmplInit, err := template.New("initConfig").Parse(injectJaegerTemplate)
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

// encodes the Envoy config so it can used
// in the init container command
func escapeYaml(yaml string) (string, error) {
	i, err := json.Marshal(yaml)
	if err != nil {
		return "", err
	}
	s := string(i)
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s, nil
}

// shared volume between the init container and Envoy
func renderJaegerConfigVolume() string {
	return configVolume
}
