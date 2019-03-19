package patch

import (
	"bufio"
	"bytes"
	"text/template"
)

const sidecarContainerTemplate = `
{
  "name": "envoy",
  "image": "{{ .ContainerImage }}",
  "securityContext": {
    "runAsUser": 1337
  },
  "ports": [
    {
      "containerPort": 9901,
      "name": "stats",
      "protocol": "TCP"
    }
  ],
  "env": [
    {
      "name": "APPMESH_VIRTUAL_NODE_NAME",
      "value": "mesh/{{ .MeshName }}/virtualNode/{{ .VirtualNodeName }}"
    },
    {
      "name": "ENVOY_LOG_LEVEL",
      "value": "{{ .LogLevel }}"
    },
    {
      "name": "AWS_REGION",
      "value": "{{ .Region }}"
    }
  ],
  "resources": {
    "requests": {
      "cpu": "{{ .CpuRequests }}",
      "memory": "{{ .MemoryRequests }}"
    }
  }
}
`

type SidecarMeta struct {
	ContainerImage  string
	MeshName        string
	VirtualNodeName string
	LogLevel        string
	Region          string
	CpuRequests     string
	MemoryRequests  string
}

func renderSidecar(meta SidecarMeta) (string, error) {
	tmpl, err := template.New("sidecar").Parse(sidecarContainerTemplate)
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
