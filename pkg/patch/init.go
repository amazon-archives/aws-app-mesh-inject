package patch

import (
	"bufio"
	"bytes"
	"text/template"
)

const initContainerTemplate = `
{
  "name": "proxyinit",
  "image": "{{ .ContainerImage }}",
  "securityContext": {
    "capabilities": {
      "add": [
        "NET_ADMIN"
      ]
    }
  },
  "env": [
    {
      "name": "APPMESH_START_ENABLED",
      "value": "1"
    },
    {
      "name": "APPMESH_IGNORE_UID",
      "value": "1337"
    },
    {
      "name": "APPMESH_ENVOY_INGRESS_PORT",
      "value": "15000"
    },
    {
      "name": "APPMESH_ENVOY_EGRESS_PORT",
      "value": "15001"
    },
    {
      "name": "APPMESH_APP_PORTS",
      "value": "{{ .Ports }}"
    },
    {
      "name": "APPMESH_EGRESS_IGNORED_IP",
      "value": "{{ .IgnoredIPs }}"
    },
    {
      "name": "APPMESH_EGRESS_IGNORED_PORTS",
      "value": "{{ .EgressIgnoredPorts }}"
    }
  ]
}
`

type InitMeta struct {
	ContainerImage     string
	Ports              string
	EgressIgnoredPorts string
	IgnoredIPs         string
}

func renderInit(meta InitMeta) (string, error) {
	tmpl, err := template.New("init").Parse(initContainerTemplate)
	if err != nil {
		return "", err
	}
	var data bytes.Buffer
	b := bufio.NewWriter(&data)

	if err := tmpl.Execute(b, meta); err != nil {
		return "", err
	}
	err = b.Flush()
	if err != nil {
		return "", err
	}
	return data.String(), nil
}
