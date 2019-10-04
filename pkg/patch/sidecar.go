package patch

import (
	"bufio"
	"bytes"
	"text/template"
)

const envoyContainerTemplate = `
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
      "name": "APPMESH_PREVIEW",
      "value": "{{ .Preview }}"
    },
    {
      "name": "ENVOY_LOG_LEVEL",
      "value": "{{ .LogLevel }}"
    }{{ if .EnableStaticConfig }},
    {
      "name": "ENVOY_STATS_CONFIG_FILE",
      "value": "/tmp/envoy/envoyconf.yaml"
    }{{ end }},
    {
      "name": "AWS_REGION",
      "value": "{{ .Region }}"
    }{{ if .InjectXraySidecar }},
    {
      "name": "ENABLE_ENVOY_XRAY_TRACING",
      "value": "1"
    }{{ end }}{{ if .EnableStatsTags }},
    {
      "name": "ENABLE_ENVOY_STATS_TAGS",
      "value": "1"
    }{{ end }}{{ if .EnableStatsD }},
    {
      "name": "ENABLE_ENVOY_DOG_STATSD",
      "value": "1"
    }{{ end }}
  ]{{ if .EnableStaticConfig }},
  "volumeMounts": [
    {
      "mountPath": "/tmp/envoy",
      "name": "config"
    }
  ]{{ end }},
  "resources": {
    "requests": {
      "cpu": "{{ .CpuRequests }}",
      "memory": "{{ .MemoryRequests }}"
    }
  }
}
`
const xrayDaemonContainerTemplate = `
{
  "name": "xray-daemon",
  "image": "amazon/aws-xray-daemon",
  "securityContext": {
    "runAsUser": 1337
  },
  "ports": [
    {
      "containerPort": 2000,
      "name": "xray",
      "protocol": "UDP"
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

const statsDExporterContainerTemplate = `
{
  "name": "statsd-exporter",
  "image": "maddox/statsd-exporter",
  "securityContext": {
    "runAsUser": 1337
  },
  "args": [
    "--web.listen-address=0.0.0.0:9102",
    "--statsd.listen-tcp=0.0.0.0:8125",
    "--statsd.listen-udp=0.0.0.0:8125"
  ],
  "ports": [
    {
      "containerPort": 9102,
      "name": "metrics",
      "protocol": "TCP"
    },
    {
      "containerPort": 8125,
      "name": "tcpin",
      "protocol": "TCP"
    },
    {
      "containerPort": 8125,
      "name": "udpin",
      "protocol": "UDP"
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
	ContainerImage              string
	MeshName                    string
	VirtualNodeName             string
	Preview                     string
	LogLevel                    string
	Region                      string
	CpuRequests                 string
	MemoryRequests              string
	EnableStaticConfig          bool
	InjectXraySidecar           bool
	EnableStatsTags             bool
	EnableStatsD                bool
	InjectStatsDExporterSidecar bool
}

func renderSidecars(meta SidecarMeta) ([]string, error) {
	var sidecars []string

	envoySidecar, err := renderTemplate("envoy", envoyContainerTemplate, meta)
	if err != nil {
		return sidecars, err
	}

	sidecars = append(sidecars, envoySidecar)

	if meta.InjectXraySidecar {
		xrayDaemonSidecar, err := renderTemplate("xray-daemon", xrayDaemonContainerTemplate, meta)
		if err != nil {
			return sidecars, err
		}

		sidecars = append(sidecars, xrayDaemonSidecar)
	}

	if meta.InjectStatsDExporterSidecar {
		statsDSidecar, err := renderTemplate("statsd-exporter", statsDExporterContainerTemplate, meta)
		if err != nil {
			return sidecars, err
		}

		sidecars = append(sidecars, statsDSidecar)
	}

	return sidecars, nil
}

func renderTemplate(name string, t string, meta SidecarMeta) (string, error) {
	tmpl, err := template.New(name).Parse(t)
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
